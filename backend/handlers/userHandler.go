package handlers

import (
	"context"
	"net/http"

	"github.com/antonbaumann/spotify-jukebox/user"
	"github.com/davecgh/go-spew/spew"
	"github.com/gorilla/mux"
	log "github.com/sirupsen/logrus"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

type UserHandler struct {
	UserCollection *mongo.Collection
}

func (h *UserHandler) addUser(user *user.Model) error {
	filter := bson.D{{"username", user.Username}}
	result := h.UserCollection.FindOne(context.TODO(), filter)
	spew.Dump(result)
	return nil
}

func (h *UserHandler) Join(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)

	newUser := user.New(vars["username"])
	if err := h.addUser(newUser); err != nil {
		log.Infof("user [%v] tried to join but username is already taken", newUser.Username)
	}

	w.WriteHeader(http.StatusOK)
	log.Println("/user/join")
}
