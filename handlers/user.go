package handlers

import (
	"errors"
	"net/http"

	"github.com/antonbaumann/spotify-jukebox/user"
	"github.com/gorilla/mux"
	log "github.com/sirupsen/logrus"
)

var (
	ErrInvalidUserRequest = errors.New("user request invalid")
)

// Join adds user to session
func (h *Handler) Join(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	username := vars["username"]

	newUser, err := user.New(username)
	if err != nil {
		log.Errorf("creating new user [%v]: %v", username, err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	if err := h.UserCollection.AddUser(newUser); err != nil {
		log.Infof("user [%v] tried to join but username is already taken", username)
		w.WriteHeader(http.StatusConflict)
		return
	}

	log.Infof("user [%v] joined", username)

	jsonResponse(w, newUser)
}

// ListUsers lists all users in the session
func (h *Handler) ListUsers(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	username := vars["username"]

	ok, err := h.validUserRequest(r)
	if err != nil {
		log.Errorf("list users: %v", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if !ok {
		log.Warnf("list users: invalid user request by [%v]", username)
		http.Error(w, ErrInvalidUserRequest.Error(), http.StatusUnauthorized)
		return
	}

	userList, err := h.UserCollection.ListUsers()
	if err != nil {
		log.Errorf("list users: %v", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	log.Infof("user [%v] requested user list", username)
	jsonResponse(w, userList)
}

func (h *Handler) validUserRequest(r *http.Request) (bool, error) {
	vars := mux.Vars(r)
	username := vars["username"]

	u, err := h.UserCollection.GetUser(username)
	// error while looking up user
	if err != nil {
		return false, err
	}
	// username does not exist
	if u == nil {
		return false, nil
	}

	secret := r.Header.Get("Authorization")
	return secret == u.Secret, nil
}
