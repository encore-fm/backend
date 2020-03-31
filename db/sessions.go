package db

import (
	"context"
	"errors"
	"fmt"

	"github.com/antonbaumann/spotify-jukebox/config"
	"github.com/antonbaumann/spotify-jukebox/session"
	"github.com/antonbaumann/spotify-jukebox/song"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readconcern"
	"go.mongodb.org/mongo-driver/mongo/readpref"
)

type SessionCollection interface {
	AddSession(ctx context.Context, sess *session.Session) error
	GetSessionByID(ctx context.Context, sessionID string) (*session.Session, error)

	GetSongByID(ctx context.Context, sessionID, songID string) (*song.Model, error)
	AddSong(ctx context.Context, sessionID string, newSong *song.Model) error
	RemoveSong(ctx context.Context, sessionID, songID string) error
	ListSongs(ctx context.Context, sessionID string) ([]*song.Model, error)
	ReplaceSong(ctx context.Context, updatedSong *song.Model) error
	Vote(ctx context.Context, songID string, username string, scoreChange float64) (*song.Model, float64, error)
}

type sessionCollection struct {
	client     *mongo.Client
	collection *mongo.Collection
}

var _ SessionCollection = (*sessionCollection)(nil)

// NewSessionCollection creates a new sessionCollection from a client
func NewSessionCollection(client *mongo.Client) SessionCollection {
	collection := client.
		Database(config.Conf.Database.DBName).
		Collection(config.Conf.Database.SessionCollectionName)
	return &sessionCollection{
		client:     client,
		collection: collection,
	}
}

// AddSession inserts a new session into session collection
// if session with this id already exists it returns `ErrSessionAlreadyExisting`
func (c *sessionCollection) AddSession(ctx context.Context, sess *session.Session) error {
	errMsg := "[db] add session: %v"
	if _, err := c.collection.InsertOne(ctx, sess); err != nil {
		if _, ok := err.(mongo.WriteException); ok {
			return fmt.Errorf(errMsg, ErrSessionAlreadyExisting)
		}
		return fmt.Errorf(errMsg, err)
	}
	return nil
}

// GetSessionByID returns a session struct if sessionID exists
// if sessionID does not exist it returns nil
// todo: write test
func (c *sessionCollection) GetSessionByID(ctx context.Context, sessionID string) (*session.Session, error) {
	errMsg := "[db] get session by id: %v"
	filter := bson.D{{"_id", sessionID}}

	var foundSess *session.Session
	err := c.collection.FindOne(ctx, filter).Decode(&foundSess)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, fmt.Errorf(errMsg, ErrNoSessionWithID)
		}
		return nil, fmt.Errorf(errMsg, err)
	}
	return foundSess, nil
}

// GetSongByID returns a song struct if songID exists
// if songID does not exist it returns nil
// todo: write test
func (c *sessionCollection) GetSongByID(ctx context.Context, sessionID string, songID string) (*song.Model, error) {
	errMsg := "[db] get song by id: %v"

	filter := bson.D{
		{"_id", sessionID},
	}
	projection := bson.D{
		{
			Key: "songList",
			Value: bson.D{
				{
					Key:   "$elemMatch",
					Value: bson.E{Key: "_id", Value: songID},
				},
			},
		},
	}

	var foundSong *song.Model
	err := c.collection.FindOne(
		ctx,
		filter,
		options.FindOne().SetProjection(projection),
	).Decode(&foundSong)

	if err == mongo.ErrNoDocuments {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf(errMsg, err)
	}
	return foundSong, nil
}

// AddSong adds a song to a session and sorts SongList
// todo: write test
func (c *sessionCollection) AddSong(ctx context.Context, sessionID string, newSong *song.Model) error {
	errMsg := "[db] add song: %v"

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
func (c *sessionCollection) RemoveSong(ctx context.Context, sessionID, songID string) error {
	errMsg := "[db] remove song: %v"
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

func (c *sessionCollection) ReplaceSong(ctx context.Context, updatedSong *song.Model) error {
	errMsg := "[db] replace song: %v"
	filter := bson.D{{"_id", updatedSong.ID}}

	result, err := c.collection.ReplaceOne(ctx, filter, updatedSong)
	if err != nil {
		return fmt.Errorf(errMsg, err)
	}
	if result.ModifiedCount != 1 {
		return fmt.Errorf(errMsg, ErrNoSongWithID)
	}
	return nil
}

// ListSongs returns a list of all songs in a session
func (c *sessionCollection) ListSongs(ctx context.Context, sessionID string) ([]*song.Model, error) {
	errMsg := "[db] list songs: %v"

	filter := bson.D{
		{"_id", sessionID},
	}
	projection := bson.D{
		{"song_list", 1},
	}

	var sessionInfo *session.Session
	err := c.collection.FindOne(
		ctx,
		filter,
		options.FindOne().SetProjection(projection),
	).Decode(&sessionInfo)

	if err == mongo.ErrNoDocuments {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf(errMsg, err)
	}
	return sessionInfo.SongList, nil
}

func (c *sessionCollection) Vote(
	ctx context.Context,
	songID string,
	username string,
	scoreChange float64,
) (*song.Model, float64, error) {
	errMsg := "[db] vote: %v"

	// start server session
	opts := options.Session().SetDefaultReadConcern(readconcern.Majority())
	sess, err := c.client.StartSession(opts)
	if err != nil {
		return nil, 0, fmt.Errorf(errMsg, err)
	}
	defer sess.EndSession(ctx)

	txnOpts := options.Transaction().SetReadPreference(readpref.Primary())
	result, err := sess.WithTransaction(ctx, func(sessCtx mongo.SessionContext) (interface{}, error) {
		songInfo, err := c.GetSongByID(sessCtx, "session_id", songID)
		if err != nil {
			_ = sessCtx.AbortTransaction(sessCtx)
			return nil, err
		}
		if songInfo == nil {
			_ = sessCtx.AbortTransaction(sessCtx)
			return nil, ErrNoSongWithID
		}
		// make sure users dont vote on songs they suggested
		if songInfo.SuggestedBy == username {
			_ = sessCtx.AbortTransaction(sessCtx)
			return nil, ErrVoteOnSuggestedSong
		}

		if scoreChange > 0 {
			// check if user did already vote up and insert if not
			if ok := songInfo.Upvoters.Add(username, scoreChange); !ok {
				_ = sessCtx.AbortTransaction(sessCtx)
				return nil, ErrUserAlreadyVoted
			}
			// if user downvoted song
			// - add score back
			// - remove user from downvoters
			if downvoteScore, ok := songInfo.Downvoters.Get(username); ok {
				scoreChange += downvoteScore
				songInfo.Downvoters.Remove(username)
			}
		}

		if scoreChange < 0 {
			// check if user did already vote down and insert if not
			if ok := songInfo.Downvoters.Add(username, -scoreChange); !ok {
				_ = sessCtx.AbortTransaction(sessCtx)
				return nil, ErrUserAlreadyVoted
			}
			// if user voted song up
			// - remove score
			// - remove user from upvoters
			if upvoteScore, ok := songInfo.Upvoters.Get(username); ok {
				scoreChange -= upvoteScore
				songInfo.Upvoters.Remove(username)
			}
		}

		// apply change and save song info in db
		songInfo.Score += scoreChange
		if err := c.ReplaceSong(sessCtx, songInfo); err != nil {
			_ = sessCtx.AbortTransaction(sessCtx)
			return nil, err
		}

		if err := sessCtx.CommitTransaction(sessCtx); err != nil {
			return nil, fmt.Errorf("commit transaction: %v", err)
		}
		return songInfo, nil
	}, txnOpts)
	if err != nil {
		return nil, 0, fmt.Errorf(errMsg, err)
	}

	songInfo, ok := result.(*song.Model)
	if !ok {
		return nil, 0, fmt.Errorf(errMsg, "cannot cast result into `*song.Model`")
	}

	return songInfo, scoreChange, nil
}
