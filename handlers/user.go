package handlers

import (
	"errors"
	"net/http"

	"github.com/antonbaumann/spotify-jukebox/song"
	"github.com/antonbaumann/spotify-jukebox/user"
	"github.com/gorilla/mux"
	log "github.com/sirupsen/logrus"
	"github.com/zmb3/spotify"
)

var ErrSpotifyNotAuthenticated = errors.New("spotify not authenticated")

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

	userList, err := h.UserCollection.ListUsers()
	if err != nil {
		log.Errorf("list users: %v", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	log.Infof("user [%v] requested user list", username)
	jsonResponse(w, userList)
}

func (h *Handler) SuggestSong(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	username := vars["username"]
	songID := vars["song_id"]

	if !h.spotifyIsAuthenticated {
		log.Errorf("suggest song: %v", ErrSpotifyNotAuthenticated)
		http.Error(w, ErrSpotifyNotAuthenticated.Error(), http.StatusInternalServerError)
		return
	}
	fullTrack, err := h.Spotify.GetTrack(spotify.ID(songID))
	if err != nil {
		log.Errorf("suggest song: retrieving info from spotify: %v", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	songInfo := song.New(username, 0, fullTrack)
	if err := h.SongCollection.AddSong(songInfo); err != nil {
		log.Errorf("suggest song: insert into songs collection: %v", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	log.Infof("suggest song: by [%v] songID [%v]", username, songID)
	jsonResponse(w, songInfo)
}

