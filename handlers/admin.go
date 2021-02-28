package handlers

import (
	"context"
	"errors"
	"net/http"

	"github.com/encore-fm/backend/db"
	"github.com/encore-fm/backend/events"
	"github.com/encore-fm/backend/session"
	"github.com/encore-fm/backend/sse"
	"github.com/encore-fm/backend/user"
	"github.com/gorilla/mux"
	log "github.com/sirupsen/logrus"
)

type AdminHandler interface {
	CreateSession(w http.ResponseWriter, r *http.Request)
	DeleteSession(w http.ResponseWriter, r *http.Request)
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

	// create admin user. contains
	// - user secret
	// - state for spotify authentication
	admin, err := user.NewAdmin(username, sess.ID)
	if err != nil {
		if errors.Is(err, user.ErrUsernameTooShort) {
			handleError(w, http.StatusBadRequest, log.DebugLevel, msg, err, UsernameTooShortError)
			return
		}
		if errors.Is(err, user.ErrUsernameTooLong) {
			handleError(w, http.StatusBadRequest, log.DebugLevel, msg, err, UsernameTooLongError)
			return
		}
		if errors.Is(err, user.ErrUsernameInvalidCharacter) {
			handleError(w, http.StatusBadRequest, log.DebugLevel, msg, err, UsernameInvalidCharacterError)
			return
		}

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
}

func (h *handler) DeleteSession(w http.ResponseWriter, r *http.Request) {
	msg := "[handler] delete session"
	ctx := context.Background()

	sessionID := r.Header.Get("Session")

	// pause spotify clients
	clients, err := h.UserCollection.GetSyncedSpotifyClients(ctx, sessionID)
	if err != nil {
		log.Errorf("%v: %v", msg, err)
	}
	for _, client := range clients {
		spotifyClient := h.spotifyAuthenticator.NewClient(client.AuthToken)
		err = spotifyClient.Pause()
		if err != nil {
			log.Errorf("%v: %v", msg, err)
		}
	}

	err = h.UserCollection.DeleteUsersBySessionID(ctx, sessionID)
	if err != nil {
		handleError(w, http.StatusInternalServerError, log.ErrorLevel, msg, err, InternalServerError)
		return
	}
	err = h.SessionCollection.DeleteSession(ctx, sessionID)
	if err != nil {
		handleError(w, http.StatusInternalServerError, log.ErrorLevel, msg, err, InternalServerError)
	}
}

func (h *handler) RemoveSong(w http.ResponseWriter, r *http.Request) {
	msg := "[handler] remove song"
	ctx := context.Background()
	vars := mux.Vars(r)
	songID := vars["song_id"]
	sessionID := r.Header.Get("Session")

	// update session time stamp
	h.SessionCollection.SetLastUpdated(ctx, sessionID)

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
