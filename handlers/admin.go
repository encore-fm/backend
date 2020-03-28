package handlers

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"

	"github.com/antonbaumann/spotify-jukebox/config"
	"github.com/antonbaumann/spotify-jukebox/sse"
	"github.com/antonbaumann/spotify-jukebox/user"
	"github.com/gorilla/mux"
	log "github.com/sirupsen/logrus"
)

var (
	ErrWrongCredentials = errors.New("username and password do not match")
)

type AdminHandler interface {
	Login(w http.ResponseWriter, r *http.Request)
	RemoveSong(w http.ResponseWriter, r *http.Request)
}

var _ AdminHandler = (*handler)(nil)

// Log in checks credentials {username, password} in request.Body
// if they match with configured admin credentials the admin-user
// struct will be returned
func (h *handler) Login(w http.ResponseWriter, r *http.Request) {
	ctx := context.Background()
	var credentials config.AdminConfig
	if err := json.NewDecoder(r.Body).Decode(&credentials); err != nil {
		log.Errorf("admin login: decoding request body: %v", err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	if credentials.Username != config.Conf.Admin.Username ||
		credentials.Password != config.Conf.Admin.Password {
		log.Errorf("admin login: %v", ErrWrongCredentials)
		http.Error(w, ErrWrongCredentials.Error(), http.StatusForbidden)
		return
	}

	admin, err := h.UserCollection.GetUser(ctx, credentials.Username)
	if err != nil {
		log.Errorf("admin login: get user from db: %v", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if admin != nil {
		log.Infof("admin login: [%v] successfully logged in", credentials.Username)
		jsonResponse(w, admin)
		return
	}

	// if user does not exist in database -> create new user
	admin, err = user.NewAdmin(credentials.Username)
	if err != nil {
		log.Errorf("admin login: create new user: %v", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if err := h.UserCollection.AddUser(ctx, admin); err != nil {
		log.Errorf("admin login: add user to db: %v", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	log.Infof("admin login: [%v] successfully logged in", credentials.Username)
	jsonResponse(w, admin)
}

func (h *handler) RemoveSong(w http.ResponseWriter, r *http.Request) {
	ctx := context.Background()
	msg := "remove song"
	vars := mux.Vars(r)
	songID := vars["song_id"]

	if err := h.SongCollection.RemoveSong(ctx, songID); err != nil {
		log.Errorf("%v: %v", msg, err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	songList, err := h.SongCollection.ListSongs(ctx)
	if err != nil {
		log.Errorf("%v: %v", msg, err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// send new playlist to broker
	event := sse.Event{
		Event: sse.PlaylistChange,
		Data:  songList,
	}
	h.Broker.Notifier <- event

	log.Infof("admin removed song [%v]", songID)
	jsonResponse(w, songList)
}
