package db

import (
	"context"
	"errors"
	"fmt"

	"github.com/antonbaumann/spotify-jukebox/config"
	"github.com/antonbaumann/spotify-jukebox/session"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

var (
	ErrSessionAlreadyExisting = errors.New("session with this id already exists")
)

type SessionCollection interface {
	AddSession(ctx context.Context, sess *session.Session) error
	GetSessionByID(ctx context.Context, sessionID string) (*session.Session, error)
}

type sessionCollection struct {
	client     *mongo.Client
	collection *mongo.Collection
}

var _ SessionCollection = (*sessionCollection)(nil)

func NewSessionCollection(client *mongo.Client) SessionCollection {
	collection := client.
		Database(config.Conf.Database.DBName).
		Collection(config.Conf.Database.SessionCollectionName)
	return &sessionCollection{
		client:     client,
		collection: collection,
	}
}

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
func (c *sessionCollection) GetSessionByID(ctx context.Context, sessionID string) (*session.Session, error) {
	errMsg := "[db] get session by id: %v"
	filter := bson.D{{"_id", sessionID}}

	var foundSess *session.Session
	err := c.collection.FindOne(ctx, filter).Decode(&foundSess)
	if err == mongo.ErrNoDocuments {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf(errMsg, err)
	}
	return foundSess, nil
}
