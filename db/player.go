package db

import (
	"context"
	"fmt"
	"github.com/antonbaumann/spotify-jukebox/config"
	"github.com/antonbaumann/spotify-jukebox/player"
	"github.com/antonbaumann/spotify-jukebox/session"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type PlayerCollection interface {
	GetPlayer(ctx context.Context, sessionID string) (*player.Player, error)
	SetPlayer(ctx context.Context, sessionID string, newPlayer *player.Player) error
	SetPaused(ctx context.Context, sessionID string, paused bool) error
	SetProgress(ctx context.Context, sessionID, progress int) error
}

type playerCollection struct {
	client     *mongo.Client
	collection *mongo.Collection
}

var _ PlayerCollection = (*playerCollection)(nil)

func NewPlayerCollection(client *mongo.Client) PlayerCollection {
	collection := client.
		Database(config.Conf.Database.DBName).
		Collection(config.Conf.Database.PlayerCollectionName)
	return &playerCollection{
		client:     client,
		collection: collection,
	}
}

func (c *playerCollection) GetPlayer(ctx context.Context, sessionID string) (*player.Player, error) {
	errMsg := "[db] get player: %w"
	filter := bson.D{{"_id", sessionID}}
	projection := bson.D{
		{"_id", 0},
		{"player", 1},
	}

	var sess session.Session
	err := c.collection.FindOne(
		ctx,
		filter,
		options.FindOne().SetProjection(projection),
	).Decode(&sess)
	if err != nil {
		return nil, fmt.Errorf(errMsg, err)
	}

	return sess.Player, nil
}

func (c *playerCollection) SetPlayer(ctx context.Context, sessionID string, newPlayer *player.Player) error {
	errMsg := "[db] set player: %w"
	filter := bson.D{{"_id", sessionID}}
	update := bson.D{
		{
			Key: "$set",
			Value: bson.D{
				{
					Key:   "player",
					Value: newPlayer,
				},
			},
		},
	}
	result, err := c.collection.UpdateOne(ctx, filter, update)
	if err != nil {
		return fmt.Errorf(errMsg, err)
	}
	if result.ModifiedCount == 0 {
		return fmt.Errorf(errMsg, ErrNoSessionWithID)
	}
	return nil
}

func (c *playerCollection) SetPaused(ctx context.Context, sessionID string, paused bool) error {
	errMsg := "[db] toggle playing: %w"
	filter := bson.D{{"_id", sessionID}}
	update := bson.D{
		{
			Key: "$set",
			Value: bson.D{
				{
					Key:   "player.paused",
					Value: paused,
				},
			},
		},
	}
	result, err := c.collection.UpdateOne(ctx, filter, update)
	if err != nil {
		return fmt.Errorf(errMsg, err)
	}
	if result.ModifiedCount == 0 {
		return fmt.Errorf(errMsg, ErrNoSessionWithID)
	}
	return nil
}

func (c *playerCollection) SetProgress(ctx context.Context, sessionID, progress int) error {
	errMsg := "[db] set progress: %w"
	filter := bson.D{{"_id", sessionID}}
	update := bson.D{
		{
			Key: "$set",
			Value: bson.D{
				{
					Key:   "player.progress",
					Value: progress,
				},
			},
		},
	}
	result, err := c.collection.UpdateOne(ctx, filter, update)
	if err != nil {
		return fmt.Errorf(errMsg, err)
	}
	if result.ModifiedCount == 0 {
		return fmt.Errorf(errMsg, ErrNoSessionWithID)
	}
	return nil
}
