package db

import (
	"context"
	"errors"

	"github.com/antonbaumann/spotify-jukebox/config"
	"github.com/antonbaumann/spotify-jukebox/user"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

var (
	ErrUsernameTaken = errors.New("requested username already taken")
)

type UserCollection struct {
	collection *mongo.Collection
}

func NewUserCollection(client *mongo.Client) *UserCollection {
	userCollection := client.
		Database(config.Conf.Database.DBName).
		Collection(config.Conf.Database.UserCollectionName)
	return &UserCollection{userCollection}
}

// Get User returns a user struct is username exists
// if username does not exist it returns nil
func (h *UserCollection) GetUser(username string) (*user.Model, error) {
	filter := bson.D{{"username", username}}
	var foundUser *user.Model
	err := h.collection.FindOne(context.TODO(), filter).Decode(&foundUser)
	if err == mongo.ErrNoDocuments {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return foundUser, nil
}

func (h *UserCollection) AddUser(newUser *user.Model) error {
	u, err := h.GetUser(newUser.Username)
	if err != nil {
		return err
	}
	if u != nil {
		return ErrUsernameTaken
	}

	_, err = h.collection.InsertOne(context.TODO(), newUser)
	return err
}

func (h *UserCollection) ListUsers() ([]*user.ListElement, error) {
	var userList []*user.ListElement
	cursor, err := h.collection.Find(context.TODO(), bson.D{{}})
	if err != nil {
		return nil, err
	}
	defer cursor.Close(context.TODO())

	for cursor.Next(context.TODO()) {
		var elem user.ListElement
		err := cursor.Decode(&elem)
		if err != nil {
			return userList, err
		}

		userList = append(userList, &elem)
	}
	return userList, nil
}
