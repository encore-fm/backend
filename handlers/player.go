package handlers

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/antonbaumann/spotify-jukebox/db"
	"github.com/antonbaumann/spotify-jukebox/events"
	"github.com/antonbaumann/spotify-jukebox/playerctrl"
	"github.com/antonbaumann/spotify-jukebox/sse"
	"github.com/antonbaumann/spotify-jukebox/user"
	"github.com/gorilla/mux"
	log "github.com/sirupsen/logrus"
)

type PlayerHandler interface {
	Play(w http.ResponseWriter, r *http.Request)
	Pause(w http.ResponseWriter, r *http.Request)
	Skip(w http.ResponseWriter, r *http.Request)
	Seek(w http.ResponseWriter, r *http.Request)
	GetState(w http.ResponseWriter, r *http.Request)
}

var _ PlayerHandler = (*handler)(nil)

func (h *handler) checkUserPermissions(w http.ResponseWriter, msg string, userID string) bool {
	msg = fmt.Sprintf("%v: check user permissions", msg)
	ctx := context.Background()

	userInfo, err := h.UserCollection.GetUserByID(ctx, userID)
	if err != nil {
		if errors.Is(err, db.ErrNoUserWithID) {
			handleError(w, http.StatusUnauthorized, log.WarnLevel, msg, err, RequestNotAuthorizedError)
			return false
		}
		handleError(w, http.StatusInternalServerError, log.ErrorLevel, msg, err, InternalServerError)
		return false
	}

	if !userInfo.IsAdmin {
		handleError(w, http.StatusUnauthorized, log.WarnLevel, msg, err, ActionNotAllowedError)
		return false
	}
	return true
}

func (h *handler) setPausedState(w http.ResponseWriter, r *http.Request, paused bool) {
	msg := "[player handler]: play / pause"

	vars := mux.Vars(r)
	username := vars["username"]
	sessionID := r.Header.Get("Session")

	if ok := h.checkUserPermissions(w, msg, user.GenerateUserID(username, sessionID)); !ok {
		return
	}

	h.eventBus.Publish(
		playerctrl.PlayPauseEvent,
		events.GroupID(sessionID),
		playerctrl.PlayPausePayload{Paused: paused},
	)
}

// Play toggles play on
func (h *handler) Play(w http.ResponseWriter, r *http.Request) {
	h.setPausedState(w, r, false)
}

// pause current song
func (h *handler) Pause(w http.ResponseWriter, r *http.Request) {
	h.setPausedState(w, r, true)
}

func (h *handler) Skip(w http.ResponseWriter, r *http.Request) {
	msg := "[player handler]: skip song"

	vars := mux.Vars(r)
	username := vars["username"]
	sessionID := r.Header.Get("Session")

	if ok := h.checkUserPermissions(w, msg, user.GenerateUserID(username, sessionID)); !ok {
		return
	}

	h.eventBus.Publish(
		playerctrl.SkipEvent,
		events.GroupID(sessionID),
		playerctrl.SkipPayload{},
	)
}

func (h *handler) Seek(w http.ResponseWriter, r *http.Request) {
	msg := "[player handler]: seek"

	vars := mux.Vars(r)
	username := vars["username"]
	sessionID := r.Header.Get("Session")

	positionMs, err := strconv.Atoi(vars["position_ms"])
	if err != nil {
		handleError(w, http.StatusBadRequest, log.WarnLevel, msg, err, RequestUrlMalformedError)
		return
	}

	if ok := h.checkUserPermissions(w, msg, user.GenerateUserID(username, sessionID)); !ok {
		return
	}

	h.eventBus.Publish(
		playerctrl.SeekEvent,
		events.GroupID(sessionID),
		playerctrl.SeekPayload{
			Progress: time.Millisecond * time.Duration(positionMs),
		},
	)
}

// todo: add component tests
// todo: add system tests
func (h *handler) GetState(w http.ResponseWriter, r *http.Request) {
	msg := "[player handler]: get state"
	ctx := context.Background()

	sessionID := r.Header.Get("Session")
	playr, err := h.PlayerCollection.GetPlayer(ctx, sessionID)
	if err != nil {
		handleError(w, http.StatusInternalServerError, log.ErrorLevel, msg, err, InternalServerError)
		return
	}

	result := &sse.PlayerStateChangePayload{
		CurrentSong: playr.CurrentSong,
		IsPlaying:   !playr.Paused,
		Progress:    playr.Progress(),
		Timestamp:   time.Now(),
	}

	jsonResponse(w, result)
}
