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

var (
	ErrSpotifyNotAuthenticated = errors.New("spotify not authenticated")
	ErrUserAlreadyVoted = errors.New("user already voted")
	ErrSongNotInCollection = errors.New("song with given ID not in db")
	ErrUserNotExisting = errors.New("user with given ID does not exist")
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

	if !h.spotifyActivated {
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

func (h *Handler) ListSongs(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	username := vars["username"]

	songList, err := h.SongCollection.ListSongs()
	if err != nil {
		log.Errorf("list users: %v", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	log.Infof("user [%v] requested song list", username)
	jsonResponse(w, songList)
}

func (h *Handler) Vote(w http.ResponseWriter, r *http.Request) {
	msg := "vote"

	vars := mux.Vars(r)
	username := vars["username"]
	songID := vars["song_id"]
	voteAction := vars["vote_action"]

	if voteAction != "up" && voteAction != "down" {
		errMsg := `vote action must be in {"up", "down"}`
		log.Errorf("%v: %v", msg, errMsg)
		http.Error(w, errMsg, http.StatusBadRequest)
		return
	}

	songInfo, err := h.SongCollection.GetSongByID(songID)
	if err != nil {
		log.Errorf("%v: %v", msg, err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if songInfo == nil {
		log.Errorf("%v: %v", msg, ErrSongNotInCollection)
		http.Error(w, ErrSongNotInCollection.Error(), http.StatusBadRequest)
		return
	}

	userInfo, err := h.UserCollection.GetUser(username)
	if err != nil {
		log.Errorf("%v: %v", msg, err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if userInfo == nil {
		log.Errorf("%v: %v", msg, ErrUserNotExisting)
		http.Error(w, ErrUserNotExisting.Error(), http.StatusBadRequest)
		return
	}

	scoreIncAmount := float64(0)

	if voteAction == "up" {
		// add user to upvoters if not in list
		upvoters, ok := songInfo.Upvoters.Add(username, userInfo.Score)
		songInfo.Upvoters = upvoters
		if !ok {
			log.Errorf("%v: %v", msg, ErrUserAlreadyVoted)
			http.Error(w, ErrUserAlreadyVoted.Error(), http.StatusBadRequest)
			return
		}
		scoreIncAmount += 1

		// remove user from downvoters if in list
		downvoters, ok := songInfo.Downvoters.Remove(username)
		songInfo.Downvoters = downvoters
		if ok {
			scoreIncAmount += 1
		}
	}
	if voteAction == "down" {
		// add user to downvoters if not in list
		downvoters, ok := songInfo.Downvoters.Add(username, userInfo.Score)
		songInfo.Downvoters = downvoters
		if !ok {
			log.Errorf("%v: %v", msg, ErrUserAlreadyVoted)
			http.Error(w, ErrUserAlreadyVoted.Error(), http.StatusBadRequest)
			return
		}
		scoreIncAmount -= 1

		// remove user from upvoters if in list
		upvoters, ok := songInfo.Upvoters.Remove(username)
		songInfo.Upvoters = upvoters
		if ok {
			scoreIncAmount -= 1
		}
	}

	// todo find good score system
	songInfo.Score += scoreIncAmount

	// write new song info to db
	if err := h.SongCollection.UpdateSong(songInfo); err != nil {
		log.Errorf("%v: %v", msg, err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// update score of user that suggested song
	if err := h.UserCollection.IncrementScore(songInfo.SuggestedBy, scoreIncAmount); err != nil {
		log.Errorf("%v: %v", msg, err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// return updated song list
	songList, err := h.SongCollection.ListSongs()
	if err != nil {
		log.Errorf("%v: %v", msg, err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	log.Infof("user [%v] %vvoted song [%v]", username, voteAction, songID)
	jsonResponse(w, songList)
}
