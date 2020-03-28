package db

import (
	"context"
	"errors"
	"fmt"

	"github.com/antonbaumann/spotify-jukebox/config"
	"github.com/antonbaumann/spotify-jukebox/song"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readconcern"
	"go.mongodb.org/mongo-driver/mongo/readpref"
)

type SongCollection interface {
	GetSongByID(ctx context.Context, songID string) (*song.Model, error)
	AddSong(ctx context.Context, newSong *song.Model) error
	RemoveSong(ctx context.Context, songID string) error
	ListSongs(ctx context.Context) ([]*song.Model, error)
	ReplaceSong(ctx context.Context, updatedSong *song.Model) error
	Vote(ctx context.Context, songID string, username string, scoreChange float64) (*song.Model, float64, error)
}

type songCollection struct {
	client     *mongo.Client
	collection *mongo.Collection
}

var _ SongCollection = (*songCollection)(nil)

var (
	ErrNoSongWithID        = errors.New("no song with this ID")
	ErrVoteOnSuggestedSong = errors.New("cannot vote on suggested song")
	ErrUserAlreadyVoted    = errors.New("user already voted")
)

func NewSongCollection(client *mongo.Client) SongCollection {
	collection := client.
		Database(config.Conf.Database.DBName).
		Collection(config.Conf.Database.SongCollectionName)
	return &songCollection{
		client:     client,
		collection: collection,
	}
}

// GetSongByID returns a song struct if songID exists
// if songID does not exist it returns nil
func (h *songCollection) GetSongByID(ctx context.Context, songID string) (*song.Model, error) {
	errMsg := "[db] get song by id: %v"
	filter := bson.D{{"_id", songID}}
	var foundSong *song.Model
	err := h.collection.FindOne(ctx, filter).Decode(&foundSong)
	if err == mongo.ErrNoDocuments {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf(errMsg, err)
	}
	return foundSong, nil
}

func (h *songCollection) AddSong(ctx context.Context, newSong *song.Model) error {
	errMsg := "[db] add song: %v"
	if _, err := h.collection.InsertOne(ctx, newSong); err != nil {
		return fmt.Errorf(errMsg, err)
	}
	return nil
}

func (h *songCollection) RemoveSong(ctx context.Context, songID string) error {
	errMsg := "[db] remove song: %v"
	filter := bson.D{{"_id", songID}}
	result, err := h.collection.DeleteOne(ctx, filter)
	if err != nil {
		return fmt.Errorf(errMsg, err)
	}
	if result.DeletedCount != 1 {
		return fmt.Errorf(errMsg, ErrNoSongWithID)
	}
	return nil
}

func (h *songCollection) ReplaceSong(ctx context.Context, updatedSong *song.Model) error {
	errMsg := "[db] replace song: %v"
	filter := bson.D{{"_id", updatedSong.ID}}

	result, err := h.collection.ReplaceOne(ctx, filter, updatedSong)
	if err != nil {
		return fmt.Errorf(errMsg, err)
	}
	if result.ModifiedCount != 1 {
		return fmt.Errorf(errMsg, ErrNoSongWithID)
	}
	return nil
}

func (h *songCollection) ListSongs(ctx context.Context) ([]*song.Model, error) {
	errMsg := "[db] list songs: %v"
	opts := options.Find()
	opts.SetSort(bson.D{
		{"score", -1},
		{"time_added", 1},
	})

	cursor, err := h.collection.Find(ctx, bson.D{}, opts)
	if err != nil {
		return nil, fmt.Errorf(errMsg, err)
	}
	defer cursor.Close(ctx)

	var songList []*song.Model
	for cursor.Next(ctx) {
		var elem song.Model
		if err := cursor.Decode(&elem); err != nil {
			return songList, fmt.Errorf(errMsg, err)
		}
		songList = append(songList, &elem)
	}
	return songList, nil
}

func (h *songCollection) Vote(
	ctx context.Context,
	songID string,
	username string,
	scoreChange float64,
) (*song.Model, float64, error) {
	errMsg := "[db] vote: %v"

	// start server session
	opts := options.Session().SetDefaultReadConcern(readconcern.Majority())
	sess, err := h.client.StartSession(opts)
	if err != nil {
		return nil, 0, fmt.Errorf(errMsg, err)
	}
	defer sess.EndSession(ctx)

	txnOpts := options.Transaction().SetReadPreference(readpref.Primary())
	result, err := sess.WithTransaction(ctx, func(sessCtx mongo.SessionContext) (interface{}, error) {
		songInfo, err := h.GetSongByID(sessCtx, songID)
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
		if err := h.ReplaceSong(sessCtx, songInfo); err != nil {
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
