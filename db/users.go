package db

import (
	"context"
	"errors"
	"fmt"

	"github.com/antonbaumann/spotify-jukebox/config"
	"github.com/antonbaumann/spotify-jukebox/user"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"golang.org/x/oauth2"
)

type UserCollection interface {
	GetUserByID(ctx context.Context, userID string) (*user.Model, error)
	GetUserByState(ctx context.Context, state string) (*user.Model, error)
	GetAdminBySessionID(ctx context.Context, sessionID string) (*user.Model, error)
	AddUser(ctx context.Context, newUser *user.Model) error
	DeleteUser(ctx context.Context, userID string) error
	DeleteUsersBySessionID(ctx context.Context, sessionID string) error
	ListUsers(ctx context.Context, sessionID string) ([]*user.ListElement, error)
	IncrementScore(ctx context.Context, username string, amount int) error
	SetToken(ctx context.Context, userID string, token *oauth2.Token) error
	SetSynchronized(ctx context.Context, userID string, synchronized bool) (*user.Model, error)
	GetSpotifyClients(ctx context.Context, sessionID string) ([]*user.SpotifyClient, error)
}

type userCollection struct {
	client     *mongo.Client
	collection *mongo.Collection
}

var _ UserCollection = (*userCollection)(nil)

func NewUserCollection(client *mongo.Client) UserCollection {
	collection := client.
		Database(config.Conf.Database.DBName).
		Collection(config.Conf.Database.UserCollectionName)
	return &userCollection{
		client:     client,
		collection: collection,
	}
}

// findOne is a wrapper around mongo's FindOne operation
// returns user if found
// if no document is found it returns nil
func (c *userCollection) findOne(ctx context.Context, filter bson.D) (*user.Model, error) {
	var foundUser *user.Model
	err := c.collection.FindOne(ctx, filter).Decode(&foundUser)
	if err != nil {
		return nil, err
	}
	return foundUser, nil
}

// Get User returns a user struct if username exists
// if username does not exist it returns nil
func (c *userCollection) GetUserByID(ctx context.Context, userID string) (*user.Model, error) {
	errMsg := "[db] get user by id : %w"
	filter := bson.D{{"_id", userID}}
	res, err := c.findOne(ctx, filter)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, fmt.Errorf(errMsg, ErrNoUserWithID)
		}
		return nil, fmt.Errorf(errMsg, err)
	}
	return res, nil
}

// Get User returns a user struct if user with `state` exists
// if `state` does not exist it returns nil
func (c *userCollection) GetUserByState(ctx context.Context, state string) (*user.Model, error) {
	errMsg := "[db] get user by state: %w"
	filter := bson.D{{"auth_state", state}}
	res, err := c.findOne(ctx, filter)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, fmt.Errorf(errMsg, ErrNoUserWithState)
		}
		return nil, fmt.Errorf(errMsg, err)
	}
	return res, nil
}

func (c *userCollection) GetAdminBySessionID(ctx context.Context, sessionID string) (*user.Model, error) {
	errMsg := "[db] get admin by sessionID: %w"
	filter := bson.D{
		{"session_id", sessionID},
		{"is_admin", true},
	}
	res, err := c.findOne(ctx, filter)
	if err != nil {
		// no admin being found with the given session id implies the session not existing
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, fmt.Errorf(errMsg, ErrNoSessionWithID)
		}
		return nil, fmt.Errorf(errMsg, err)
	}
	return res, nil
}

func (c *userCollection) AddUser(ctx context.Context, newUser *user.Model) error {
	errMsg := "[db] add user: %w"
	if _, err := c.collection.InsertOne(ctx, newUser); err != nil {
		if _, ok := err.(mongo.WriteException); ok {
			return fmt.Errorf(errMsg, ErrUsernameTaken)
		}
		return fmt.Errorf(errMsg, err)
	}
	return nil
}

func (c *userCollection) DeleteUser(ctx context.Context, userID string) error {
	errMsg := "[db] delete user: %w"

	filter := bson.M{"_id": userID}
	res, err := c.collection.DeleteOne(ctx, filter)
	if err != nil {
		return fmt.Errorf(errMsg, err)
	}
	if res.DeletedCount == 0 {
		return fmt.Errorf(errMsg, ErrNoUserWithID)
	}

	return nil
}

func (c *userCollection) DeleteUsersBySessionID(ctx context.Context, sessionID string) error {
	errMsg := "[db] delete users by session id: %w"

	filter := bson.M{"session_id": sessionID}
	res, err := c.collection.DeleteMany(ctx, filter)
	if err != nil {
		return fmt.Errorf(errMsg, err)
	}
	if res.DeletedCount == 0 {
		return fmt.Errorf(errMsg, ErrNoSessionWithID)
	}

	return nil
}

func (c *userCollection) ListUsers(ctx context.Context, sessionID string) ([]*user.ListElement, error) {
	errMsg := "[db] list users: %w"
	var userList []*user.ListElement
	cursor, err := c.collection.Find(ctx, bson.D{{"session_id", sessionID}})
	if err != nil {
		return nil, fmt.Errorf(errMsg, err)
	}
	defer cursor.Close(ctx)

	for cursor.Next(ctx) {
		var elem user.ListElement
		err := cursor.Decode(&elem)
		if err != nil {
			return userList, fmt.Errorf(errMsg, err)
		}

		userList = append(userList, &elem)
	}
	return userList, nil
}

func (c *userCollection) IncrementScore(ctx context.Context, userID string, amount int) error {
	errMsg := "[db] increment user score: %w"

	filter := bson.D{{"_id", userID}}
	update := bson.D{
		{
			Key: "$inc",
			Value: bson.D{
				{
					Key:   "score",
					Value: amount,
				},
			},
		},
	}
	result, err := c.collection.UpdateOne(ctx, filter, update)
	if err != nil {
		return fmt.Errorf(errMsg, err)
	}
	if result.ModifiedCount == 0 {
		return fmt.Errorf(errMsg, ErrNoUserWithID)
	}
	return nil
}

// Set token
// - sets spotify authorization token
// - sets spotify_authorized field to true
func (c *userCollection) SetToken(ctx context.Context, userID string, token *oauth2.Token) error {
	errMsg := "[db] set token: %w"
	filter := bson.D{{"_id", userID}}
	update := bson.D{
		{
			Key: "$set",
			Value: bson.D{
				{
					Key:   "auth_token",
					Value: token,
				},
				{
					Key:   "spotify_authorized",
					Value: true,
				},
			},
		},
	}
	result, err := c.collection.UpdateOne(ctx, filter, update)
	if err != nil {
		return fmt.Errorf(errMsg, err)
	}
	if result.ModifiedCount == 0 {
		return fmt.Errorf(errMsg, ErrNoUserWithID)
	}
	return nil
}

func (c *userCollection) SetSynchronized(ctx context.Context, userID string, synchronized bool) (*user.Model, error) {
	errMsg := "[db] set synchronized: %w"
	filter := bson.M{
		"_id":                userID,
		"spotify_authorized": true,
	}
	update := bson.M{
		"$set": bson.M{"spotify_synchronized": synchronized},
	}
	var usr *user.Model

	after := options.After
	opt := options.FindOneAndUpdateOptions{ReturnDocument: &after}
	err := c.collection.FindOneAndUpdate(ctx, filter, update, &opt).Decode(&usr)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, fmt.Errorf(errMsg, ErrNoUserWithID)
		}
		return nil, fmt.Errorf(errMsg, err)
	}

	return usr, nil
}

func (c *userCollection) GetSpotifyClients(ctx context.Context, sessionID string) ([]*user.SpotifyClient, error) {
	errMsg := "[db] get spotify clients: %w"
	filter := bson.D{
		{"session_id", sessionID},
		{"spotify_authorized", true},
		{"spotify_synchronized", true},
	}
	projection := bson.D{
		{"_id", 1},
		{"username", 1},
		{"session_id", 1},
		{"is_admin", 1},
		{"auth_token", 1},
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

	var clients []*user.SpotifyClient
	for cursor.Next(ctx) {
		var client user.SpotifyClient
		err := cursor.Decode(&client)
		if err != nil {
			return nil, fmt.Errorf(errMsg, err)
		}

		clients = append(clients, &client)
	}

	return clients, nil
}
