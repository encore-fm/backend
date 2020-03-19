package handlers

import (
	"context"
	"errors"
	"net/http"

	"github.com/antonbaumann/spotify-jukebox/user"
	"github.com/gorilla/mux"
	log "github.com/sirupsen/logrus"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

type UserHandler struct {
	UserCollection *mongo.Collection
}

var (
	ErrUsernameTaken = errors.New("requested username already taken")
)

func (h *UserHandler) getUser(username string) (*user.Model, error) {
	filter := bson.D{{"username", username}}
	var foundUser *user.Model
	err := h.UserCollection.FindOne(context.TODO(), filter).Decode(&foundUser)
	// if username does not exist -> no documents in result error
	if err != nil {
		return nil, err
	}
	return foundUser, nil
}

func (h *UserHandler) addUser(newUser *user.Model) error {
	_, err := h.getUser(newUser.Username)
	if err != mongo.ErrNoDocuments {
		return ErrUsernameTaken
	}

	_, err = h.UserCollection.InsertOne(context.TODO(), newUser)
	return err
}

// Join adds user to session
func (h *UserHandler) Join(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	username := vars["username"]

	newUser, err := user.New(username)
	if err != nil {
		log.Errorf("creating new user [%v]: %v", username, err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	if err := h.addUser(newUser); err != nil {
		log.Infof("user [%v] tried to join but username is already taken", username)
		w.WriteHeader(http.StatusConflict)
		return
	}

	log.Infof("user [%v] joined", username)

	jsonResponse(w, newUser)
}
