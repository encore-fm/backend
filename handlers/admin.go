package handlers

import (
	"context"
	"encoding/json"
	"errors"
	"github.com/antonbaumann/spotify-jukebox/playerctrl"
	"net/http"
	"time"

	"github.com/antonbaumann/spotify-jukebox/db"
	"github.com/antonbaumann/spotify-jukebox/events"
	"github.com/antonbaumann/spotify-jukebox/session"
	"github.com/antonbaumann/spotify-jukebox/sse"
	"github.com/antonbaumann/spotify-jukebox/user"
	"github.com/gorilla/mux"
	log "github.com/sirupsen/logrus"
)

type AdminHandler interface {
	CreateSession(w http.ResponseWriter, r *http.Request)
	RemoveSong(w http.ResponseWriter, r *http.Request)
	PutPlayerState(w http.ResponseWriter, r *http.Request)
}

var _ AdminHandler = (*handler)(nil)

func (h *handler) CreateSession(w http.ResponseWriter, r *http.Request) {
	msg := "[handler] create session"
	ctx := context.Background()

	vars := mux.Vars(r)
	username := vars["username"]

	// create new session (contains random session id)
	sess, err := session.New()
	if err != nil {
		handleError(w, http.StatusInternalServerError, log.ErrorLevel, msg, err, InternalServerError)
		return
	}

	// save session in db
	err = h.SessionCollection.AddSession(ctx, sess)
	if err != nil {
		if errors.Is(err, db.ErrSessionAlreadyExisting) {
			handleError(w, http.StatusConflict, log.WarnLevel, msg, err, SessionConflictError)
		} else {
			handleError(w, http.StatusInternalServerError, log.ErrorLevel, msg, err, InternalServerError)
		}
		return
	}

	// create admin user. contains
	// - user secret
	// - state for spotify authentication
	admin, err := user.NewAdmin(username, sess.ID)
	if err != nil {
		handleError(w, http.StatusInternalServerError, log.ErrorLevel, msg, err, InternalServerError)
		return
	}

	// save admin user in db
	if err := h.UserCollection.AddUser(ctx, admin); err != nil {
		if errors.Is(err, db.ErrUsernameTaken) {
			handleError(w, http.StatusConflict, log.WarnLevel, msg, err, UserConflictError)
		} else {
			handleError(w, http.StatusInternalServerError, log.ErrorLevel, msg, err, InternalServerError)
		}
		return
	}

	// create authentication url containing auth state
	// auth state will later be used to link spotify callback to user
	authUrl := h.spotifyAuthenticator.AuthURLWithDialog(admin.AuthState)

	response := &struct {
		UserInfo *user.Model `json:"user_info"`
		AuthUrl  string      `json:"auth_url"`
	}{
		UserInfo: admin,
		AuthUrl:  authUrl,
	}

	log.Infof("%v: [%v] successfully created session with id [%v]", msg, username, sess.ID)
	jsonResponse(w, response)

	// register session at playerController
	h.eventBus.Publish(
		playerctrl.RegisterSessionEvent,
		events.GroupIDAny,
		playerctrl.RegisterSessionPayload{SessionID: sess.ID},
	)
}

func (h *handler) RemoveSong(w http.ResponseWriter, r *http.Request) {
	ctx := context.Background()
	msg := "[handler] remove song"
	vars := mux.Vars(r)
	songID := vars["song_id"]
	sessionID := r.Header.Get("Session")

	if err := h.SongCollection.RemoveSong(ctx, sessionID, songID); err != nil {
		if errors.Is(err, db.ErrNoSessionWithID) {
			handleError(w, http.StatusBadRequest, log.WarnLevel, msg, err, SessionNotFoundError)
		} else if errors.Is(err, db.ErrNoSongWithID) {
			handleError(w, http.StatusBadRequest, log.WarnLevel, msg, err, SongNotFoundError)
		} else {
			handleError(w, http.StatusInternalServerError, log.ErrorLevel, msg, err, InternalServerError)
		}
		return
	}

	songList, err := h.SongCollection.ListSongs(ctx, sessionID)
	if err != nil {
		if errors.Is(err, db.ErrNoSessionWithID) {
			handleError(w, http.StatusBadRequest, log.WarnLevel, msg, err, SessionNotFoundError)
		} else {
			handleError(w, http.StatusInternalServerError, log.ErrorLevel, msg, err, InternalServerError)
		}
		return
	}

	h.eventBus.Publish(sse.PlaylistChange, events.GroupID(sessionID), songList)

	log.Infof("%v: admin removed song [%v]", msg, songID)
	jsonResponse(w, songList)
}

func (h *handler) PutPlayerState(w http.ResponseWriter, r *http.Request) {
	msg := "[handler] player state change"

	sessionID := r.Header.Get("Session")

	var payload playerctrl.StateChangedPayload
	err := json.NewDecoder(r.Body).Decode(&payload)
	if err != nil || payload.SongID == "" {
		handleError(w, http.StatusBadRequest, log.WarnLevel, msg, err, RequestBodyMalformedError)
		return
	}

	h.eventBus.Publish(playerctrl.AdminStateChangedEvent, events.GroupID(sessionID), payload)

	response := struct {
		Message string `json:"message"`
	}{
		Message: "success",
	}
	jsonResponse(w, response)
}

// todo: think about splitting this in two seperate play/pause endpoints for simplicity reasons
func (h *handler) SetPlaying(w http.ResponseWriter, r *http.Request) {
	msg := "[handler] set playing"
	ctx := context.Background()

	sessionID := r.Header.Get("Session")

	// todo have parameter in request url or in body?
	var payload struct {
		Playing bool `json:"playing"`
	}
	err := json.NewDecoder(r.Body).Decode(&payload)
	if err != nil {
		handleError(w, http.StatusBadRequest, log.WarnLevel, msg, err, RequestBodyMalformedError)
		return
	}
	playing := payload.Playing

	// todo notify clients

	// update session player
	if err := h.SessionCollection.SetPaused(ctx, sessionID, !playing); err != nil {
		if errors.Is(err, db.ErrNoSessionWithID) {
			handleError(w, http.StatusBadRequest, log.WarnLevel, msg, err, SessionNotFoundError)
		} else {
			handleError(w, http.StatusInternalServerError, log.ErrorLevel, msg, err, InternalServerError)
		}
		return
	}

	response := struct {
		Message string `json:"message"`
	}{
		Message: "success",
	}

	log.Infof("%v: admin played", msg)
	jsonResponse(w, response)
}

func (h *handler) Skip(w http.ResponseWriter, r *http.Request) {
	msg := "[handler] skip"
	ctx := context.Background()

	sessionID := r.Header.Get("Session")

	// todo have parameter in request url or in body?
	var payload struct {
		SeekTime time.Duration `json:"seek_time"`
	}
	err := json.NewDecoder(r.Body).Decode(&payload)
	if err != nil {
		handleError(w, http.StatusBadRequest, log.WarnLevel, msg, err, RequestBodyMalformedError)
		return
	}
	seekTime := payload.SeekTime

	// todo seek player in db and notify clients

}

func (h *handler) Seek(w http.ResponseWriter, r *http.Request) {
	msg := "[handler] seek"
	ctx := context.Background()

	sessionID := r.Header.Get("Session")

	// todo get current playing song, see if request is valid and notify clients
	// todo atomicity??
}
