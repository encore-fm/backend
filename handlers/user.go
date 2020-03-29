package handlers

import (
	"context"
	"errors"
	"math"
	"net/http"

	"github.com/antonbaumann/spotify-jukebox/song"
	"github.com/antonbaumann/spotify-jukebox/sse"
	"github.com/antonbaumann/spotify-jukebox/user"
	"github.com/gorilla/mux"
	log "github.com/sirupsen/logrus"
	"github.com/zmb3/spotify"
)

var (
	ErrSpotifyNotAuthenticated = errors.New("spotify not authenticated")
	ErrSongNotInCollection     = errors.New("song with given ID not in db")
	ErrUserNotExisting         = errors.New("user with given ID does not exist")
	ErrSessionNotExisting      = errors.New("session with given ID does not exist")
)

type UserHandler interface {
	Join(w http.ResponseWriter, r *http.Request)
	ListUsers(w http.ResponseWriter, r *http.Request)
	SuggestSong(w http.ResponseWriter, r *http.Request)
	ListSongs(w http.ResponseWriter, r *http.Request)
	Vote(w http.ResponseWriter, r *http.Request)
}

var _ UserHandler = (*handler)(nil)

// Join adds user to session
// - check if session with this id exists
// - create new user and save in db
// - create auth url
func (h *handler) Join(w http.ResponseWriter, r *http.Request) {
	msg := "[handler] join"
	ctx := context.Background()

	vars := mux.Vars(r)
	username := vars["username"]
	sessionID := vars["session_id"]

	// check if session with given id exists
	sess, err := h.SessionCollection.GetSessionByID(ctx, sessionID)
	if err != nil {
		log.Errorf("%v: %v", msg, err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if sess == nil {
		log.Warnf("%v: %v", msg, err)
		http.Error(w, ErrSessionNotExisting.Error(), http.StatusBadRequest)
		return
	}

	newUser, err := user.New(username, sessionID)
	if err != nil {
		log.Errorf("%v: creating new user [%v]: %v", msg, username, err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// save user in db
	if err := h.UserCollection.AddUser(ctx, newUser); err != nil {
		log.Errorf("%v: save user: %v", msg, err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// create authentication url containing auth state
	// auth state will later be used to link spotify callback to user
	authUrl := h.spotifyAuthenticator.AuthURLWithDialog(newUser.AuthState)

	response := &struct {
		UserInfo *user.Model `json:"user_info"`
		AuthUrl  string      `json:"auth_url"`
	}{
		UserInfo: newUser,
		AuthUrl:  authUrl,
	}

	log.Infof("%v: [%v] successfully joined session with id [%v]", msg, username, sess.ID)
	jsonResponse(w, response)
}

// ListUsers lists all users in the session
func (h *handler) ListUsers(w http.ResponseWriter, r *http.Request) {
	ctx := context.Background()
	vars := mux.Vars(r)
	username := vars["username"]

	userList, err := h.UserCollection.ListUsers(ctx)
	if err != nil {
		log.Errorf("list users: %v", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	log.Infof("user [%v] requested user list", username)
	jsonResponse(w, userList)
}

func (h *handler) SuggestSong(w http.ResponseWriter, r *http.Request) {
	ctx := context.Background()
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
	if err := h.SongCollection.AddSong(ctx, songInfo); err != nil {
		log.Errorf("suggest song: insert into songs collection: %v", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	log.Infof("suggest song: by [%v] songID [%v]", username, songID)
	jsonResponse(w, songInfo)

	// fetch songList and send event
	songList, err := h.SongCollection.ListSongs(ctx)
	if err != nil {
		log.Errorf("suggest song: event: %v", err)
	}
	// send new playlist to broker
	event := sse.Event{
		Event: sse.PlaylistChange,
		Data:  songList,
	}
	h.Broker.Notifier <- event
}

func (h *handler) ListSongs(w http.ResponseWriter, r *http.Request) {
	ctx := context.Background()

	msg := "[handler] list songs"
	vars := mux.Vars(r)
	username := vars["username"]

	songList, err := h.SongCollection.ListSongs(ctx)
	if err != nil {
		log.Errorf("%v: %v", msg, err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	log.Infof("user [%v] requested song list", username)
	jsonResponse(w, songList)
}

func (h *handler) Vote(w http.ResponseWriter, r *http.Request) {
	ctx := context.Background()

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

	userInfo, err := h.UserCollection.GetUserByUsername(ctx, username)
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

	scoreChange := math.Max(userInfo.Score, 1)
	if voteAction == "down" {
		scoreChange = -scoreChange
	}

	songInfo, change, err := h.SongCollection.Vote(ctx, songID, username, scoreChange)
	if err != nil {
		log.Errorf("%v: %v", msg, err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// update score of user that suggested song
	if err := h.UserCollection.IncrementScore(ctx, songInfo.SuggestedBy, change); err != nil {
		log.Errorf("%v: %v", msg, err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// return updated song list
	songList, err := h.SongCollection.ListSongs(ctx)
	if err != nil {
		log.Errorf("%v: %v", msg, err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	log.Infof("user [%v] %vvoted song [%v]", username, voteAction, songID)
	jsonResponse(w, songList)

	// send new playlist to broker
	event := sse.Event{
		Event: sse.PlaylistChange,
		Data:  songList,
	}
	h.Broker.Notifier <- event
}
