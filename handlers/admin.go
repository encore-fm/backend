package handlers

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/antonbaumann/spotify-jukebox/config"
	"github.com/antonbaumann/spotify-jukebox/user"
	"github.com/gorilla/mux"
	log "github.com/sirupsen/logrus"
)

var (
	ErrUserNotAdmin  = errors.New("user not an admin")
	ErrWrongPassword = errors.New("wrong password")
)

// Log in checks credentials {username, password} in request.Body
// if they match with configured admin credentials the admin-user
// struct will be returned
func (h *Handler) Login(w http.ResponseWriter, r *http.Request) {
	var credentials config.AdminConfig
	if err := json.NewDecoder(r.Body).Decode(&credentials); err != nil {
		log.Errorf("admin login: decoding request body: %v", err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	if credentials.Username != config.Conf.Admin.Username {
		log.Errorf("admin login: %v", ErrUserNotAdmin)
		http.Error(w, ErrUserNotAdmin.Error(), http.StatusForbidden)
		return
	}
	if credentials.Password != config.Conf.Admin.Password {
		log.Errorf("admin login: %v", ErrWrongPassword)
		http.Error(w, ErrWrongPassword.Error(), http.StatusForbidden)
		return
	}

	admin, err := h.UserCollection.GetUser(credentials.Username)
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

	if err := h.UserCollection.AddUser(admin); err != nil {
		log.Errorf("admin login: add user to db: %v", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	log.Infof("admin login: [%v] successfully logged in", credentials.Username)
	jsonResponse(w, admin)
}

func (h *Handler) RemoveSong(w http.ResponseWriter, r *http.Request) {
	msg := "remove song"
	vars := mux.Vars(r)
	songID := vars["song_id"]

	if err := h.SongCollection.RemoveSong(songID); err != nil {
		log.Errorf("%v: %v", msg, err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	songList, err := h.SongCollection.ListSongs()
	if err != nil {
		log.Errorf("%v: %v", msg, err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	log.Infof("admin removed song [%v]", songID)
	jsonResponse(w, songList)
}