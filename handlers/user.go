package handlers

import (
	"context"
	"errors"
	"github.com/antonbaumann/spotify-jukebox/db"
	"github.com/antonbaumann/spotify-jukebox/song"
	"github.com/antonbaumann/spotify-jukebox/sse"
	"github.com/antonbaumann/spotify-jukebox/user"
	"github.com/gorilla/mux"
	log "github.com/sirupsen/logrus"
	"github.com/zmb3/spotify"
	"net/http"
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
		if errors.Is(err, db.ErrNoSessionWithID) {
			handleError(w, http.StatusNotFound, log.ErrorLevel, msg, err, SessionNotFoundError)
		} else {
			handleError(w, http.StatusInternalServerError, log.ErrorLevel, msg, err, InternalServerError)
		}
		return
	}

	newUser, err := user.New(username, sessionID)
	if err != nil {
		handleError(w, http.StatusInternalServerError, log.ErrorLevel, msg, err, InternalServerError)
		return
	}

	// save user in db
	if err := h.UserCollection.AddUser(ctx, newUser); err != nil {
		if errors.Is(err, db.ErrUsernameTaken) {
			handleError(w, http.StatusConflict, log.ErrorLevel, msg, err, UserConflictError)
		} else {
			handleError(w, http.StatusInternalServerError, log.ErrorLevel, msg, err, InternalServerError)
		}
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
	msg := "[handler] list users"
	ctx := context.Background()

	vars := mux.Vars(r)
	username := vars["username"]
	sessionID := r.Header.Get("Session")

	userList, err := h.UserCollection.ListUsers(ctx, sessionID)
	if err != nil {
		handleError(w, http.StatusInternalServerError, log.ErrorLevel, msg, err, InternalServerError)
		return
	}

	log.Infof("user [%v] requested user list", username)
	jsonResponse(w, userList)
}

// ListUsers add's a new song to the song_list
func (h *handler) SuggestSong(w http.ResponseWriter, r *http.Request) {
	msg := "[handler] suggest song"
	ctx := context.Background()

	vars := mux.Vars(r)
	username := vars["username"]
	songID := vars["song_id"]
	sessionID := r.Header.Get("Session")

	fullTrack, err := h.Spotify.GetTrack(spotify.ID(songID))
	if err != nil {
		// todo: should mostly be UserError -> better checks
		handleError(w, http.StatusInternalServerError, log.ErrorLevel, msg, err, InternalServerError)
		return
	}

	// if user suggest's song he automatically votes up
	songInfo := song.New(username, 1, fullTrack)
	songInfo.Upvoters = append(songInfo.Upvoters, username)

	if err := h.SessionCollection.AddSong(ctx, sessionID, songInfo); err != nil {
		if errors.Is(err, db.ErrSongAlreadyInSession) {
			handleError(w, http.StatusConflict, log.WarnLevel, msg, err, SongConflictError)
		} else {
			handleError(w, http.StatusInternalServerError, log.ErrorLevel, msg, err, InternalServerError)
		}
		return
	}

	log.Infof("%v: by [%v] songID [%v]", msg, username, songID)
	jsonResponse(w, songInfo)

	// fetch songList and send event
	songList, err := h.SessionCollection.ListSongs(ctx, sessionID)
	if err != nil {
		log.Errorf("%v: event: %v", msg, err)
	}
	// send new playlist to broker
	event := sse.Event{
		Event: sse.PlaylistChange,
		Data:  songList,
	}
	h.Broker.Notifier <- event
}

// ListSongs returns all songs in one session
func (h *handler) ListSongs(w http.ResponseWriter, r *http.Request) {
	msg := "[handler] list songs"
	ctx := context.Background()

	vars := mux.Vars(r)
	username := vars["username"]

	sessionID := r.Header.Get("Session")

	songList, err := h.SessionCollection.ListSongs(ctx, sessionID)
	if err != nil {
		handleError(w, http.StatusInternalServerError, log.ErrorLevel, msg, err, InternalServerError)
		return
	}

	log.Infof("%v: user [%v]", msg, username)
	jsonResponse(w, songList)
}

// Vote handles user votes on songs
func (h *handler) Vote(w http.ResponseWriter, r *http.Request) {
	msg := "[handler] vote"
	ctx := context.Background()

	vars := mux.Vars(r)
	username := vars["username"]
	songID := vars["song_id"]
	voteAction := vars["vote_action"]

	// get session id from headers
	sessionID := r.Header.Get("Session")

	if voteAction != "up" && voteAction != "down" {
		handleError(w, http.StatusBadRequest, log.ErrorLevel, msg, ErrBadVoteAction, BadVoteError)
		return
	}

	// the scoreChange of the song score has to be applied to the user score
	var scoreChange int
	var err error

	if voteAction == "up" {
		scoreChange, err = h.SessionCollection.VoteUp(ctx, sessionID, songID, username)
		if err != nil {
			handleError(w, http.StatusInternalServerError, log.ErrorLevel, msg, err, InternalServerError)
			return
		}
	} else {
		scoreChange, err = h.SessionCollection.VoteDown(ctx, sessionID, songID, username)
		if err != nil {
			handleError(w, http.StatusInternalServerError, log.ErrorLevel, msg, err, InternalServerError)
			return
		}
	}

	songInfo, err := h.SessionCollection.GetSongByID(ctx, sessionID, songID)
	if err != nil {
		handleError(w, http.StatusInternalServerError, log.ErrorLevel, msg, err, InternalServerError)
		return
	}

	// if user updates his vote on own song -> dont update his score
	if songInfo.SuggestedBy != username {
		// update score of user that suggested song
		if err := h.UserCollection.IncrementScore(
			ctx,
			user.GenerateUserID(songInfo.SuggestedBy, sessionID),
			scoreChange,
		); err != nil {
			handleError(w, http.StatusInternalServerError, log.ErrorLevel, msg, err, InternalServerError)
			return
		}
	}

	// return updated song list
	songList, err := h.SessionCollection.ListSongs(ctx, sessionID)
	if err != nil {
		handleError(w, http.StatusInternalServerError, log.ErrorLevel, msg, err, InternalServerError)
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
