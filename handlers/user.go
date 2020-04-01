package handlers

import (
	"context"
	"errors"
	"github.com/antonbaumann/spotify-jukebox/db"
	"math"
	"net/http"

	"github.com/antonbaumann/spotify-jukebox/song"
	"github.com/antonbaumann/spotify-jukebox/sse"
	"github.com/antonbaumann/spotify-jukebox/user"
	"github.com/gorilla/mux"
	log "github.com/sirupsen/logrus"
	"github.com/zmb3/spotify"
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
			HandleError(w, http.StatusNotFound, log.ErrorLevel, msg, err, SessionNotFoundError)
		} else {
			HandleError(w, http.StatusInternalServerError, log.ErrorLevel, msg, err, InternalServerError)
		}
		return
	}

	newUser, err := user.New(username, sessionID)
	if err != nil {
		HandleError(w, http.StatusInternalServerError, log.ErrorLevel, msg, err, InternalServerError)
		return
	}

	// save user in db
	if err := h.UserCollection.AddUser(ctx, newUser); err != nil {
		if errors.Is(err, db.ErrUsernameTaken) {
			HandleError(w, http.StatusConflict, log.ErrorLevel, msg, err, UserConflictError)
		} else {
			HandleError(w, http.StatusInternalServerError, log.ErrorLevel, msg, err, InternalServerError)
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

	userList, err := h.UserCollection.ListUsers(ctx)
	if err != nil {
		HandleError(w, http.StatusInternalServerError, log.ErrorLevel, msg, err, InternalServerError)
		return
	}

	log.Infof("user [%v] requested user list", username)
	jsonResponse(w, userList)
}

func (h *handler) SuggestSong(w http.ResponseWriter, r *http.Request) {
	msg := "[handler] suggest song"
	ctx := context.Background()

	vars := mux.Vars(r)
	username := vars["username"]
	songID := vars["song_id"]

	if !h.spotifyActivated {
		HandleError(w, http.StatusInternalServerError, log.ErrorLevel, msg, ErrSpotifyNotAuthenticated, InternalServerError)
		return
	}
	fullTrack, err := h.Spotify.GetTrack(spotify.ID(songID))
	if err != nil {
		HandleError(w, http.StatusInternalServerError, log.ErrorLevel, msg, err, InternalServerError)
		return
	}

	songInfo := song.New(username, 0, fullTrack)
	if err := h.SongCollection.AddSong(ctx, songInfo); err != nil {
		if errors.Is(err, db.ErrSongAlreadySuggested) {
			HandleError(w, http.StatusConflict, log.ErrorLevel, msg, err, SongConflictError)
		} else {
			HandleError(w, http.StatusInternalServerError, log.ErrorLevel, msg, err, InternalServerError)
		}
		return
	}

	log.Infof("%v: by [%v] songID [%v]", msg, username, songID)
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
	msg := "[handler] list songs"
	ctx := context.Background()

	vars := mux.Vars(r)
	username := vars["username"]

	songList, err := h.SongCollection.ListSongs(ctx)
	if err != nil {
		HandleError(w, http.StatusInternalServerError, log.ErrorLevel, msg, err, InternalServerError)
		return
	}

	log.Infof("%v: user [%v]", msg, username)
	jsonResponse(w, songList)
}

func (h *handler) Vote(w http.ResponseWriter, r *http.Request) {
	msg := "[handler] vote"
	ctx := context.Background()

	vars := mux.Vars(r)
	username := vars["username"]
	songID := vars["song_id"]
	voteAction := vars["vote_action"]

	// get session id from headers
	sessID := r.Header.Get("Session")

	if voteAction != "up" && voteAction != "down" {
		HandleError(w, http.StatusBadRequest, log.ErrorLevel, msg, ErrBadVoteAction, BadVoteError)
		return
	}

	userInfo, err := h.UserCollection.GetUserByID(ctx, user.GenerateUserID(username, sessID))
	if err != nil {
		if errors.Is(err, db.ErrNoUserWithID) {
			HandleError(w, http.StatusNotFound, log.ErrorLevel, msg, err, UserNotFoundError)
		} else {
			HandleError(w, http.StatusInternalServerError, log.ErrorLevel, msg, err, InternalServerError)
		}
		return
	}

	scoreChange := math.Max(userInfo.Score, 1)
	if voteAction == "down" {
		scoreChange = -scoreChange
	}

	songInfo, change, err := h.SongCollection.Vote(ctx, songID, username, scoreChange)
	if err != nil {
		if errors.Is(err, db.ErrVoteOnSuggestedSong) {
			HandleError(w, http.StatusConflict, log.ErrorLevel, msg, err, VoteOnSuggestedSongError)
		} else if errors.Is(err, db.ErrUserAlreadyVoted) {
			HandleError(w, http.StatusConflict, log.ErrorLevel, msg, err, UserAlreadyVotedError)
		} else {
			HandleError(w, http.StatusInternalServerError, log.ErrorLevel, msg, err, InternalServerError)
		}
		return
	}

	// update score of user that suggested song
	if err := h.UserCollection.IncrementScore(ctx, songInfo.SuggestedBy, change); err != nil {
		HandleError(w, http.StatusInternalServerError, log.ErrorLevel, msg, err, InternalServerError)
		return
	}

	// return updated song list
	songList, err := h.SongCollection.ListSongs(ctx)
	if err != nil {
		HandleError(w, http.StatusInternalServerError, log.ErrorLevel, msg, err, InternalServerError)
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
