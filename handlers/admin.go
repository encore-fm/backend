package handlers

import (
	"context"
	"errors"
	"net/http"

	"github.com/antonbaumann/spotify-jukebox/db"
	"github.com/antonbaumann/spotify-jukebox/events"
	"github.com/antonbaumann/spotify-jukebox/player"
	"github.com/antonbaumann/spotify-jukebox/session"
	"github.com/antonbaumann/spotify-jukebox/sse"
	"github.com/antonbaumann/spotify-jukebox/user"
	"github.com/gorilla/mux"
	log "github.com/sirupsen/logrus"
)

type AdminHandler interface {
	CreateSession(w http.ResponseWriter, r *http.Request)
	RemoveSong(w http.ResponseWriter, r *http.Request)
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
		player.RegisterSessionEvent,
		events.GroupIDAny,
		player.RegisterSessionPayload{SessionID: sess.ID},
	)
}

func (h *handler) RemoveSong(w http.ResponseWriter, r *http.Request) {
	ctx := context.Background()
	msg := "[handler] remove song"
	vars := mux.Vars(r)
	songID := vars["song_id"]
	sessionID := r.Header.Get("Session")

	if err := h.SessionCollection.RemoveSong(ctx, sessionID, songID); err != nil {
		if errors.Is(err, db.ErrNoSessionWithID) {
			handleError(w, http.StatusBadRequest, log.WarnLevel, msg, err, SessionNotFoundError)
		} else if errors.Is(err, db.ErrNoSongWithID) {
			handleError(w, http.StatusBadRequest, log.WarnLevel, msg, err, SongNotFoundError)
		} else {
			handleError(w, http.StatusInternalServerError, log.ErrorLevel, msg, err, InternalServerError)
		}
		return
	}

	songList, err := h.SessionCollection.ListSongs(ctx, sessionID)
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
