package db

import (
	"context"
	"errors"
	"fmt"

	"github.com/antonbaumann/spotify-jukebox/config"
	"github.com/antonbaumann/spotify-jukebox/user"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

var (
	ErrUsernameTaken              = errors.New("requested username already taken")
	ErrIncrementScoreNoUserWithID = errors.New("increment score: no user with given ID")
)

type UserCollection interface {
	GetUser(username string) (*user.Model, error)
	AddUser(newUser *user.Model) error
	ListUsers() ([]*user.ListElement, error)
	IncrementScore(username string, amount float64) error
}

type userCollection struct {
	collection *mongo.Collection
}

var _ UserCollection = (*userCollection)(nil)

func NewUserCollection(client *mongo.Client) UserCollection {
	collection := client.
		Database(config.Conf.Database.DBName).
		Collection(config.Conf.Database.UserCollectionName)
	return &userCollection{collection}
}

// Get User returns a user struct is username exists
// if username does not exist it returns nil
func (h *userCollection) GetUser(username string) (*user.Model, error) {
	errMsg := "get user: %v"
	filter := bson.D{{"username", username}}
	var foundUser *user.Model
	err := h.collection.FindOne(context.TODO(), filter).Decode(&foundUser)
	if err == mongo.ErrNoDocuments {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf(errMsg, err)
	}
	return foundUser, nil
}

func (h *userCollection) AddUser(newUser *user.Model) error {
	errMsg := "add user: %v"
	u, err := h.GetUser(newUser.Username)
	if err != nil {
		return fmt.Errorf(errMsg, err)
	}
	if u != nil {
		return fmt.Errorf(errMsg, ErrUsernameTaken)
	}

	if _, err = h.collection.InsertOne(context.TODO(), newUser); err != nil {
		return fmt.Errorf(errMsg, err)
	}
	return nil
}

func (h *userCollection) ListUsers() ([]*user.ListElement, error) {
	errMsg := "list users: %v"
	var userList []*user.ListElement
	cursor, err := h.collection.Find(context.TODO(), bson.D{{}})
	if err != nil {
		return nil, fmt.Errorf(errMsg, err)
	}
	defer cursor.Close(context.TODO())

	for cursor.Next(context.TODO()) {
		var elem user.ListElement
		err := cursor.Decode(&elem)
		if err != nil {
			return userList, fmt.Errorf(errMsg, err)
		}

		userList = append(userList, &elem)
	}
	return userList, nil
}

func (h *userCollection) IncrementScore(username string, amount float64) error {
	errMsg := "increment user score: %v"
	filter := bson.D{{"username", username}}
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
	result, err := h.collection.UpdateOne(context.TODO(), filter, update)
	if err != nil {
		return fmt.Errorf(errMsg, err)
	}
	if result.ModifiedCount != 1 {
		return fmt.Errorf(errMsg, ErrIncrementScoreNoUserWithID)
	}
	return nil
}
