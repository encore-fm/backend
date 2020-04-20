package db

import (
	"context"
	"fmt"
	"time"

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
	SetPaused(ctx context.Context, sessionID string) error
	SetPlaying(ctx context.Context, sessionID string) error
	IncrementProgress(ctx context.Context, sessionID string, progress time.Duration) error
}

type playerCollection struct {
	client     *mongo.Client
	collection *mongo.Collection
}

var _ PlayerCollection = (*playerCollection)(nil)

func NewPlayerCollection(client *mongo.Client) PlayerCollection {
	collection := client.
		Database(config.Conf.Database.DBName).
		Collection(config.Conf.Database.SessionCollectionName)
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
	if result.MatchedCount == 0 {
		return fmt.Errorf(errMsg, ErrNoSessionWithID)
	}
	return nil
}

func (c *playerCollection) SetPaused(ctx context.Context, sessionID string) error {
	errMsg := "[db] set paused: %w"
	filter := bson.D{
		{"_id", sessionID},
		{"player.paused", false},
	}
	update := bson.D{
		{
			Key: "$set",
			Value: bson.D{
				{
					Key:   "player.paused",
					Value: true,
				},
				{
					Key:   "player.pause_start",
					Value: time.Now(),
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
	return nil
}

func (c *playerCollection) SetPlaying(ctx context.Context, sessionID string) error {
	errMsg := "[db] set playing: %w"
	filter := bson.D{
		{"_id", sessionID},
		{"player.paused", true},
	}

	update := bson.A{
		bson.D{
			{
				Key: "$set",
				Value: bson.D{
					{
						Key:   "player.paused",
						Value: false,
					},
					{
						Key: "player.pause_duration",
						Value: bson.D{
							{
								"$add",
								bson.A{
									"$player.pause_duration",
									bson.D{
										{
											Key: "$multiply", // convert milliseconds to nanoseconds
											Value: bson.A{
												1000000,
												bson.D{
													{
														Key:   "$subtract",
														Value: bson.A{time.Now(), "$player.pause_start"},
													},
												},
											},
										},
									},
								},
							},
						},
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
	return nil
}

func (c *playerCollection) IncrementProgress(ctx context.Context, sessionID string, progress time.Duration) error {
	errMsg := "[db] increment progress: %w"
	filter := bson.D{{"_id", sessionID}}
	update := bson.D{
		{
			Key: "$inc",
			Value: bson.D{
				{
					Key:   "player.pause_duration",
					Value: progress.Nanoseconds(),
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
	return nil
}
