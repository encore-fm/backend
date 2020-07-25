package db

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/antonbaumann/spotify-jukebox/config"
	"github.com/antonbaumann/spotify-jukebox/session"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"

	log "github.com/sirupsen/logrus"
)

type SessionCollection interface {
	AddSession(ctx context.Context, sess *session.Session) error
	DeleteSession(ctx context.Context, sessionID string) error
	GetSessionByID(ctx context.Context, sessionID string) (*session.Session, error)
	ListSessionIDs(ctx context.Context) ([]string, error)
	ListExpiredSessions(ctx context.Context, sessionExpiration time.Duration) ([]string, error)
	DeleteSessions(ctx context.Context, sessionIDs []string) error
	SetLastUpdated(ctx context.Context, sessionID string)
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

// DeleteSession deletes a session from the session collection
// if session with this id does not exists, returns `ErrNoSessionWithID` error
func (c *sessionCollection) DeleteSession(ctx context.Context, sessionID string) error {
	errMsg := "[db] delete session: %w"

	filter := bson.M{"_id": sessionID}
	res, err := c.collection.DeleteOne(ctx, filter)
	if err != nil {
		return fmt.Errorf(errMsg, err)
	}
	if res.DeletedCount == 0 {
		return fmt.Errorf(errMsg, ErrNoSessionWithID)
	}

	return nil
}

// deletes multiple sessions simultaneously
func (c *sessionCollection) DeleteSessions(ctx context.Context, sessionIDs []string) error {
	errMsg := "[db] delete sessions: %w"
	filter := bson.M{
		"_id": bson.M{
			"$in": sessionIDs,
		},
	}
	res, err := c.collection.DeleteMany(ctx, filter)
	if err != nil {
		return fmt.Errorf(errMsg, err)
	}
	if res.DeletedCount != int64(len(sessionIDs)) {
		deleteErr := errors.New(fmt.Sprintf("one or more sessions with the given ids could not be deleted, "+
			"expected count: %v, got: %v", len(sessionIDs), res.DeletedCount))
		return fmt.Errorf(errMsg, deleteErr)
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

func (c *sessionCollection) ListExpiredSessions(ctx context.Context, sessionExpiration time.Duration) ([]string, error) {
	errMsg := "[db] list expired sessions: %w"

	expirationDate := time.Now().Add(-sessionExpiration)
	filter := bson.M{
		"last_updated": bson.M{
			"$lt": expirationDate,
		},
	}
	projection := bson.M{
		"_id": 1,
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
		sessID := &struct {
			SessID string `bson:"_id"`
		}{}
		err := cursor.Decode(sessID)
		if err != nil {
			return nil, fmt.Errorf(errMsg, err)
		}
		sessIDs = append(sessIDs, sessID.SessID)
	}

	return sessIDs, nil
}

// sets the last updated timestamp of the specified session to time.Now()
func (c *sessionCollection) SetLastUpdated(ctx context.Context, sessionID string) {
	errMsg := "[db] refresh session: %w"

	filter := bson.D{{"_id", sessionID}}
	update := bson.D{
		{
			Key: "$set",
			Value: bson.D{
				{
					Key:   "last_updated",
					Value: time.Now(),
				},
			},
		},
	}

	result, err := c.collection.UpdateOne(ctx, filter, update)
	if err != nil {
		log.Errorf(errMsg, err)
	}
	if result.MatchedCount == 0 {
		log.Errorf(errMsg, ErrNoSessionWithID)
	}
}
