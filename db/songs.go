package db

import (
	"context"
	"fmt"

	"github.com/antonbaumann/spotify-jukebox/config"
	"github.com/antonbaumann/spotify-jukebox/session"
	"github.com/antonbaumann/spotify-jukebox/song"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type SongCollection interface {
	GetSongByID(ctx context.Context, sessionID, songID string) (*song.Model, error)
	AddSong(ctx context.Context, sessionID string, newSong *song.Model) error
	RemoveSong(ctx context.Context, sessionID, songID string) error
	ListSongs(ctx context.Context, sessionID string) ([]*song.Model, error)
	VoteUp(ctx context.Context, sessionID, songID, username string) (int, error)
	VoteDown(ctx context.Context, sessionID, songID, username string) (int, error)
}

type songCollection struct {
	client     *mongo.Client
	collection *mongo.Collection
}

var _ SongCollection = (*songCollection)(nil)

// NewSessionCollection creates a new sessionCollection from a client
func NewSongCollection(client *mongo.Client) SongCollection {
	collection := client.
		Database(config.Conf.Database.DBName).
		Collection(config.Conf.Database.SessionCollectionName)
	return &songCollection{
		client:     client,
		collection: collection,
	}
}

// GetSongByID returns a song struct if songID exists
// if songID does not exist it returns nil
// todo: write test
func (c *songCollection) GetSongByID(ctx context.Context, sessionID string, songID string) (*song.Model, error) {

	errMsg := "[db] get song by id: %w"

	filter := bson.D{
		{"_id", sessionID},
		{"song_list.id", songID},
	}
	projection := bson.D{
		{"_id", 0},
		{"song_list.$", 1},
	}

	var sess *session.Session
	err := c.collection.FindOne(
		ctx,
		filter,
		options.FindOne().SetProjection(projection),
	).Decode(&sess)
	if err != nil {
		return nil, fmt.Errorf(errMsg, err)
	}

	if len(sess.SongList) == 0 {
		return nil, fmt.Errorf(errMsg, ErrNoSongWithID)
	}

	return sess.SongList[0], nil
}

// AddSong adds a song to a session and sorts SongList
// Errors:
// - ErrSongAlreadyInSession
func (c *songCollection) AddSong(ctx context.Context, sessionID string, newSong *song.Model) error {
	errMsg := "[db] add song: %w"

	filter := bson.D{
		{"_id", sessionID},
		{
			"song_list.id",
			bson.D{
				{
					"$ne",
					newSong.ID,
				},
			},
		},
	}

	// todo: dont sort on insert
	// or sort on insert and update but dont on find
	update := bson.D{
		{
			"$push",
			bson.D{
				{
					"song_list",
					bson.D{
						{"$each", []*song.Model{newSong}},
						{
							"$sort",
							bson.D{
								{"score", -1},
								{"time_added", 1},
							},
						},
					},
				},
			},
		},
	}

	result, err := c.collection.UpdateOne(ctx, filter, update)
	if err != nil {
		if _, ok := err.(mongo.WriteException); ok {
			return fmt.Errorf(errMsg, ErrSongAlreadyInSession)
		}
		return fmt.Errorf(errMsg, err)
	}

	if result.MatchedCount == 0 {
		return fmt.Errorf(errMsg, ErrSongAlreadyInSession)
	}
	return nil
}

// RemoveSong removes a song from collection
// todo: write test
func (c *songCollection) RemoveSong(ctx context.Context, sessionID, songID string) error {
	errMsg := "[db] remove song: %w"
	filter := bson.D{{"_id", sessionID}}
	update := bson.D{
		{
			Key: "$pull",
			Value: bson.D{
				{
					Key: "song_list",
					Value: bson.D{
						{"id", songID},
					},
				},
			},
		},
	}

	result, err := c.collection.UpdateOne(ctx, filter, update)
	if err != nil {
		return fmt.Errorf(errMsg, err)
	}
	if result.MatchedCount == 0 {
		return fmt.Errorf(errMsg, ErrNoSessionWithID)
	}
	if result.ModifiedCount == 0 {
		return fmt.Errorf(errMsg, ErrNoSongWithID)
	}
	return nil
}

// ListSongs returns a sorted list of all songs in a session
// todo: test
func (c *songCollection) ListSongs(ctx context.Context, sessionID string) ([]*song.Model, error) {
	errMsg := "[db] list songs: %w"

	matchStage := bson.D{
		{"$match", bson.D{{"_id", sessionID}}},
	}

	unwindStage := bson.D{
		{"$unwind", "$song_list"},
	}

	sortStage := bson.D{
		{
			"$sort",
			bson.D{
				{"song_list.score", -1},
				{"song_list.time_added", 1},
			},
		},
	}

	projectStage := bson.D{
		{"$project", bson.D{
			{"_id", 1},
			{"song_list", 1},
		}},
	}

	cursor, err := c.collection.Aggregate(
		ctx,
		mongo.Pipeline{matchStage, unwindStage, sortStage, projectStage},
	)
	if err != nil {
		return nil, fmt.Errorf(errMsg, err)
	}
	defer cursor.Close(ctx)

	songList := make([]*song.Model, 0)

	// define custom result struct because I'm too stupid
	// to make the mongo aggregation pipeline do what i want
	type result struct {
		Song *song.Model `bson:"song_list"`
	}

	for cursor.Next(ctx) {
		var res result
		if err := cursor.Decode(&res); err != nil {
			return nil, fmt.Errorf(errMsg, err)
		}
		songList = append(songList, res.Song)
	}

	return songList, nil
}

func (c *songCollection) VoteUp(
	ctx context.Context,
	sessionID string,
	songID string,
	username string,
) (int, error) {
	errMsg := "[db] vote up: %w"

	// case 1: 	user not in Upvoters && user not in Downvoters
	//		   	-> add user to Upvoters
	// 			-> increment score by 1
	// case 2: 	user in Upvoters 	&& user not in Downvoters
	//		   	-> remove user from Upvoters
	// 			-> decrement score by 1
	// case 3: 	user not in Upvoters && user in Downvoters
	//		   	-> remove user from Downvoters
	// 			-> add user to Upvoters
	// 			-> increment score by 2

	// case 1: filters for
	// - _id: sessionID
	// - song_list must contain a song with
	// 		- id = songID
	//		- user not in upvoters
	//      - user not in downvoters
	scoreChange := +1
	filter := bson.D{
		{"_id", sessionID},
		{
			"song_list",
			bson.D{
				{
					"$elemMatch",
					bson.D{
						{"id", songID},
						{"upvoters", bson.D{{Key: "$ne", Value: username}}},
						{"downvoters", bson.D{{Key: "$ne", Value: username}}},
					},
				},
			},
		},
	}
	update := bson.D{
		{
			"$inc",
			bson.D{{"song_list.$[song].score", 1}},
		},
		{
			"$push",
			bson.D{{"song_list.$[song].upvoters", username}},
		},
	}
	opts := options.Update().SetArrayFilters(options.ArrayFilters{
		Filters: []interface{}{bson.M{"song.id": songID}},
	})
	result, err := c.collection.UpdateOne(ctx, filter, update, opts)
	if err != nil {
		return 0, fmt.Errorf(errMsg, err)
	}
	// check if modified
	if result.ModifiedCount > 0 {
		return scoreChange, nil
	}

	// case 2: filters for
	// - _id: sessionID
	// - song_list must contain a song with
	// 		- id = songID
	//		- user in upvoters
	//      - user not in downvoters
	scoreChange = -1
	filter = bson.D{
		{"_id", sessionID},
		{
			"song_list",
			bson.D{
				{
					"$elemMatch",
					bson.D{
						{"id", songID},
						{"upvoters", bson.D{{Key: "$eq", Value: username}}},
						{"downvoters", bson.D{{Key: "$ne", Value: username}}},
					},
				},
			},
		},
	}
	update = bson.D{
		{
			"$inc",
			bson.D{{"song_list.$[song].score", scoreChange}},
		},
		{
			"$pull",
			bson.D{{"song_list.$[song].upvoters", username}},
		},
	}
	opts = options.Update().SetArrayFilters(options.ArrayFilters{
		Filters: []interface{}{bson.M{"song.id": songID}},
	})
	result, err = c.collection.UpdateOne(ctx, filter, update, opts)
	if err != nil {
		return 0, fmt.Errorf(errMsg, err)
	}
	// check if modified
	if result.ModifiedCount > 0 {
		return scoreChange, nil
	}

	// case 3: filters for
	// - _id: sessionID
	// - song_list must contain a song with
	// 		- id = songID
	//		- user not in upvoters
	//      - user in downvoters
	scoreChange = +2
	filter = bson.D{
		{"_id", sessionID},
		{
			"song_list",
			bson.D{
				{
					"$elemMatch",
					bson.D{
						{"id", songID},
						{"upvoters", bson.D{{Key: "$ne", Value: username}}},
						{"downvoters", bson.D{{Key: "$eq", Value: username}}},
					},
				},
			},
		},
	}
	update = bson.D{
		{
			"$inc",
			bson.D{{"song_list.$[song].score", scoreChange}},
		},
		{
			"$pull",
			bson.D{{"song_list.$[song].downvoters", username}},
		},
		{
			"$push",
			bson.D{{"song_list.$[song].upvoters", username}},
		},
	}
	opts = options.Update().SetArrayFilters(options.ArrayFilters{
		Filters: []interface{}{bson.M{"song.id": songID}},
	})
	result, err = c.collection.UpdateOne(ctx, filter, update, opts)
	if err != nil {
		return 0, fmt.Errorf(errMsg, err)
	}
	// check if modified
	if result.ModifiedCount > 0 {
		return scoreChange, nil
	}

	return 0, fmt.Errorf(errMsg, ErrIllegalState)
}

func (c *songCollection) VoteDown(
	ctx context.Context,
	sessionID string,
	songID string,
	username string,
) (int, error) {
	errMsg := "[db] vote down: %w"

	// case 1: 	user not in Upvoters && user not in Downvoters
	//		   	-> add user to Downvoters
	// 			-> decrement score by 1
	// case 2: 	user not in Upvoters && user in Downvoters
	//		   	-> remove user from Downvoters
	// 			-> increment score by 1
	// case 3: 	user in Upvoters && user not in Downvoters
	//		   	-> remove user from Upvoters
	// 			-> add user to Downvoters
	// 			-> decrement score by 2

	// case 1: filters for
	// - _id: sessionID
	// - song_list must contain a song with
	// 		- id = songID
	//		- user not in upvoters
	//      - user not in downvoters
	scoreChange := -1
	filter := bson.D{
		{"_id", sessionID},
		{
			"song_list",
			bson.D{
				{
					"$elemMatch",
					bson.D{
						{"id", songID},
						{"upvoters", bson.D{{Key: "$ne", Value: username}}},
						{"downvoters", bson.D{{Key: "$ne", Value: username}}},
					},
				},
			},
		},
	}
	update := bson.D{
		{
			"$inc",
			bson.D{{"song_list.$[song].score", scoreChange}},
		},
		{
			"$push",
			bson.D{{"song_list.$[song].downvoters", username}},
		},
	}
	opts := options.Update().SetArrayFilters(options.ArrayFilters{
		Filters: []interface{}{bson.M{"song.id": songID}},
	})
	result, err := c.collection.UpdateOne(ctx, filter, update, opts)
	if err != nil {
		return 0, fmt.Errorf(errMsg, err)
	}
	// check if modified
	if result.ModifiedCount > 0 {
		return scoreChange, nil
	}

	// case 2: filters for
	// - _id: sessionID
	// - song_list must contain a song with
	// 		- id = songID
	//		- user not in upvoters
	//      - user in downvoters
	scoreChange = +1
	filter = bson.D{
		{"_id", sessionID},
		{
			"song_list",
			bson.D{
				{
					"$elemMatch",
					bson.D{
						{"id", songID},
						{"upvoters", bson.D{{Key: "$ne", Value: username}}},
						{"downvoters", bson.D{{Key: "$eq", Value: username}}},
					},
				},
			},
		},
	}
	update = bson.D{
		{
			"$inc",
			bson.D{{"song_list.$[song].score", scoreChange}},
		},
		{
			"$pull",
			bson.D{{"song_list.$[song].downvoters", username}},
		},
	}
	opts = options.Update().SetArrayFilters(options.ArrayFilters{
		Filters: []interface{}{bson.M{"song.id": songID}},
	})
	result, err = c.collection.UpdateOne(ctx, filter, update, opts)
	if err != nil {
		return 0, fmt.Errorf(errMsg, err)
	}
	// check if modified
	if result.ModifiedCount > 0 {
		return scoreChange, nil
	}

	// case 3: filters for
	// - _id: sessionID
	// - song_list must contain a song with
	// 		- id = songID
	//		- user in upvoters
	//      - user not in downvoters
	scoreChange = -2
	filter = bson.D{
		{"_id", sessionID},
		{
			"song_list",
			bson.D{
				{
					"$elemMatch",
					bson.D{
						{"id", songID},
						{"upvoters", bson.D{{Key: "$eq", Value: username}}},
						{"downvoters", bson.D{{Key: "$ne", Value: username}}},
					},
				},
			},
		},
	}
	update = bson.D{
		{
			"$inc",
			bson.D{{"song_list.$[song].score", scoreChange}},
		},
		{
			"$pull",
			bson.D{{"song_list.$[song].upvoters", username}},
		},
		{
			"$push",
			bson.D{{"song_list.$[song].downvoters", username}},
		},
	}
	opts = options.Update().SetArrayFilters(options.ArrayFilters{
		Filters: []interface{}{bson.M{"song.id": songID}},
	})
	result, err = c.collection.UpdateOne(ctx, filter, update, opts)
	if err != nil {
		return 0, fmt.Errorf(errMsg, err)
	}
	// check if modified
	if result.ModifiedCount > 0 {
		return scoreChange, nil
	}

	return 0, fmt.Errorf(errMsg, ErrIllegalState)
}
