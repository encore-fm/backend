package handlers

import (
	"net/http"

	"github.com/encore-fm/backend/events"
	"github.com/encore-fm/backend/playerctrl"
	"github.com/gorilla/mux"
)

type DebugHandler interface {
	ResetControllerState(w http.ResponseWriter, r *http.Request)
}

var _ DebugHandler = (*handler)(nil)

func (h *handler) ResetControllerState(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	sessionID := vars["session_id"]

	h.eventBus.Publish(
		playerctrl.ResetEvent,
		events.GroupID(sessionID),
		playerctrl.ResetPayload{
			SessionID: sessionID,
		},
	)
}
