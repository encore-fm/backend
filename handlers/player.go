package handlers

import (
	"context"
	"errors"
	"net/http"

	"github.com/antonbaumann/spotify-jukebox/db"
	"github.com/antonbaumann/spotify-jukebox/events"
	"github.com/antonbaumann/spotify-jukebox/playerctrl"
	"github.com/antonbaumann/spotify-jukebox/user"
	"github.com/gorilla/mux"
	log "github.com/sirupsen/logrus"
)

type PlayerHandler interface {
	Play(w http.ResponseWriter, r *http.Request)
	Pause(w http.ResponseWriter, r *http.Request)
	//Skip(w http.ResponseWriter, r *http.Request)
	//Seek(w http.ResponseWriter, r *http.Request)
}

var _ PlayerHandler = (*handler)(nil)

func (h *handler) setPausedState(w http.ResponseWriter, r *http.Request, paused bool) {
	msg := "[player handler]: play / pause"
	ctx := context.Background()

	vars := mux.Vars(r)
	username := vars["username"]
	sessionID := r.Header.Get("Session")

	userInfo, err := h.UserCollection.GetUserByID(ctx, user.GenerateUserID(username, sessionID))
	if err != nil {
		if errors.Is(err, db.ErrNoUserWithID) {
			handleError(w, http.StatusUnauthorized, log.WarnLevel, msg, err, RequestNotAuthorizedError)
			return
		}
		handleError(w, http.StatusInternalServerError, log.ErrorLevel, msg, err, InternalServerError)
		return
	}

	if !userInfo.IsAdmin {
		handleError(w, http.StatusUnauthorized, log.WarnLevel, msg, err, ActionNotAllowedError)
		return
	}

	h.eventBus.Publish(
		playerctrl.PlayPauseEvent,
		events.GroupID(sessionID),
		playerctrl.PlayPausePayload{Paused: paused},
	)
}

func (h *handler) Play(w http.ResponseWriter, r *http.Request) {
	h.setPausedState(w, r, false)
}

func (h *handler) Pause(w http.ResponseWriter, r *http.Request) {
	h.setPausedState(w, r, true)
}
