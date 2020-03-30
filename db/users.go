package db

import (
	"context"
	"errors"
	"fmt"

	"github.com/antonbaumann/spotify-jukebox/config"
	"github.com/antonbaumann/spotify-jukebox/user"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"golang.org/x/oauth2"
)

var (
	ErrUsernameTaken              = errors.New("requested username already taken")
	ErrIncrementScoreNoUserWithID = errors.New("increment score: no user with given ID")
	ErrNothingModified            = errors.New("nothing modified")
)

type UserCollection interface {
	GetUserByID(ctx context.Context, userID string) (*user.Model, error)
	GetUserByState(ctx context.Context, username string) (*user.Model, error)
	AddUser(ctx context.Context, newUser *user.Model) error
	ListUsers(ctx context.Context, ) ([]*user.ListElement, error)
	IncrementScore(ctx context.Context, username string, amount float64) error
	SetToken(ctx context.Context, username string, token *oauth2.Token) error
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
func (h *userCollection) findOne(ctx context.Context, filter bson.D) (*user.Model, error) {
	var foundUser *user.Model
	err := h.collection.FindOne(ctx, filter).Decode(&foundUser)
	if err == mongo.ErrNoDocuments {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return foundUser, nil
}

// Get User returns a user struct if username exists
// if username does not exist it returns nil
func (h *userCollection) GetUserByID(ctx context.Context, userID string) (*user.Model, error) {
	errMsg := "[db] get user by username : %v"
	filter := bson.D{{"_id", userID}}
	res, err := h.findOne(ctx, filter)
	if err != nil {
		return nil, fmt.Errorf(errMsg, err)
	}
	return res, nil
}

// Get User returns a user struct if user with `state` exists
// if `state` does not exist it returns nil
func (h *userCollection) GetUserByState(ctx context.Context, state string) (*user.Model, error) {
	errMsg := "[db] get user by state: %v"
	filter := bson.D{{"auth_state", state}}
	res, err := h.findOne(ctx, filter)
	if err != nil {
		return nil, fmt.Errorf(errMsg, err)
	}
	return res, nil
}

func (h *userCollection) AddUser(ctx context.Context, newUser *user.Model) error {
	errMsg := "[db] add user: %v"
	if _, err := h.collection.InsertOne(ctx, newUser); err != nil {
		if _, ok := err.(mongo.WriteException); ok {
			return fmt.Errorf(errMsg, ErrUsernameTaken)
		}
		return fmt.Errorf(errMsg, err)
	}
	return nil
}

func (h *userCollection) ListUsers(ctx context.Context) ([]*user.ListElement, error) {
	errMsg := "[db] list users: %v"
	var userList []*user.ListElement
	cursor, err := h.collection.Find(ctx, bson.D{{}})
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

func (h *userCollection) IncrementScore(ctx context.Context, username string, amount float64) error {
	errMsg := "[db] increment user score: %v"
	filter := bson.D{{"_id", username}}
	update := bson.D{
		bson.E{
			Key: "$inc",
			Value: bson.D{
				bson.E{
					Key:   "score",
					Value: amount,
				},
			},
		},
	}
	result, err := h.collection.UpdateOne(ctx, filter, update)
	if err != nil {
		return fmt.Errorf(errMsg, err)
	}
	if result.ModifiedCount != 1 {
		return fmt.Errorf(errMsg, ErrIncrementScoreNoUserWithID)
	}
	return nil
}

// Set token
// - sets spotify authorization token
// - sets spotify_authorized field to true
func (h *userCollection) SetToken(ctx context.Context, userID string, token *oauth2.Token) error {
	errMsg := "[db] set token: %v"
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
	result, err := h.collection.UpdateOne(ctx, filter, update)
	if err != nil {
		return fmt.Errorf(errMsg, err)
	}
	if result.ModifiedCount == 0 {
		return fmt.Errorf(errMsg, ErrNothingModified)
	}
	return nil
}
