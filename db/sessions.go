package db

import (
	"context"
	"errors"
	"fmt"
	"github.com/antonbaumann/spotify-jukebox/config"
	"github.com/antonbaumann/spotify-jukebox/session"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"

	log "github.com/sirupsen/logrus"
)

type SessionCollection interface {
	AddSession(ctx context.Context, sess *session.Session) error
	GetSessionByID(ctx context.Context, sessionID string) (*session.Session, error)
	ListSessionIDs(ctx context.Context) ([]string, error)
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
	errMsg := "[db] add session: %w"
	if _, err := c.collection.InsertOne(ctx, sess); err != nil {
		if err, ok := err.(mongo.WriteException); ok {
			log.Error(err)
			return fmt.Errorf(errMsg, ErrSessionAlreadyExisting)
		}
		return fmt.Errorf(errMsg, err)
	}
	return nil
}

// GetSessionByID returns a session struct if sessionID exists
// if sessionID does not exist it returns ErrNoSessionWithID
// todo: write test
func (c *sessionCollection) GetSessionByID(ctx context.Context, sessionID string) (*session.Session, error) {
	errMsg := "[db] get session by id: %w"
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

func (c *sessionCollection) ListSessionIDs(ctx context.Context) ([]string, error) {
	errMsg := "[db] get sessionID list: %w"
	filter := bson.D{}
	projection := bson.D{
		{"song_list", 0},
	}

	cursor, err := c.collection.Find(
		ctx,
		filter,
		options.Find().SetProjection(projection),
	)
	if err != nil {
		return nil, fmt.Errorf(errMsg, err)
	}
	defer cursor.Close(ctx)

	var sessIDs []string
	for cursor.Next(ctx) {
		var sess session.Session
		err := cursor.Decode(&sess)
		if err != nil {
			return nil, fmt.Errorf(errMsg, err)
		}
		sessIDs = append(sessIDs, sess.ID)
	}

	return sessIDs, nil
}
