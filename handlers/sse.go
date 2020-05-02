package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/antonbaumann/spotify-jukebox/playerctrl"
	"github.com/antonbaumann/spotify-jukebox/user"
	"net/http"
	"time"

	"github.com/antonbaumann/spotify-jukebox/events"
	"github.com/antonbaumann/spotify-jukebox/player"
	"github.com/antonbaumann/spotify-jukebox/sse"
	"github.com/gorilla/mux"
	log "github.com/sirupsen/logrus"
)

type SSEHandler interface {
	ServeHTTP(w http.ResponseWriter, r *http.Request)
}

var _ SSEHandler = (*handler)(nil)

// This Broker method handles and HTTP request at the "/events/{username}/{session_id}" URL.
func (h *handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	msg := "[sse] serve http: %v"

	vars := mux.Vars(r)
	sessionID := vars["session_id"]
	username := vars["username"]
	userID := user.GenerateUserID(username, sessionID)

	// Make sure that the writer supports flushing.
	//
	f, ok := w.(http.Flusher)
	if !ok {
		http.Error(w, "Streaming unsupported!", http.StatusInternalServerError)
		return
	}
	// subscribe to changes
	sub := h.eventBus.Subscribe(
		[]events.EventType{sse.PlaylistChange, sse.PlayerStateChange, sse.UserListChange},
		[]events.GroupID{events.GroupID(sessionID)},
	)

	// register active sse connection
	_, err := h.UserCollection.AddSSEConnection(ctx, userID)
	if err != nil {
		log.Errorf("%v: %v", msg, err)
	}
	// publish an sse connection established event to sync user
	h.eventBus.Publish(
		playerctrl.SSEConnectionEvent,
		events.GroupID(sessionID),
		playerctrl.SSEConnectionPayload{UserID: userID, ConnectionEstablished: true},
	)

	// Listen to the closing of the http connection
	go func() {
		<-ctx.Done()
		// Remove this client from the map of attached clients
		// when `EventHandler` exits.

		h.eventBus.Unsubscribe(sub)

		// unregister sse connection
		numberOfConnections, err := h.UserCollection.RemoveSSEConnection(context.Background(), userID)
		if err != nil {
			log.Errorf("%v: %v", msg, err)
		}
		// desynchronize user if no more connections are active
		if numberOfConnections == 0 {
			h.eventBus.Publish(
				playerctrl.SSEConnectionEvent,
				events.GroupID(sessionID),
				playerctrl.SSEConnectionPayload{UserID: userID, ConnectionEstablished: false},
			)
		}

		log.Info("[sse] HTTP connection just closed")
	}()

	// Set the headers related to event streaming.
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	w.Header().Set("Transfer-Encoding", "chunked")

	playr, err := h.PlayerCollection.GetPlayer(ctx, sessionID)
	if err != nil {
		log.Errorf("%v: %v", msg, err)
	}

	// return empty player if player is nil
	if playr == nil {
		playr = player.New()
	}

	playlist, err := h.SongCollection.ListSongs(ctx, sessionID)
	if err != nil {
		log.Errorf("%v: %v", msg, err)
	}

	userList, err := h.UserCollection.ListUsers(ctx, sessionID)
	if err != nil {
		log.Errorf("%v: %v", msg, err)
	}

	playerState := sse.PlayerStateChangePayload{
		CurrentSong: playr.CurrentSong,
		IsPlaying:   !playr.Paused,
		ProgressMs:  playr.Progress().Milliseconds(),
		Timestamp:   time.Now(),
	}

	sendEvent(w, f, msg, sse.PlayerStateChange, events.GroupID(sessionID), playerState)
	sendEvent(w, f, msg, sse.PlaylistChange, events.GroupID(sessionID), playlist)
	sendEvent(w, f, msg, sse.UserListChange, events.GroupID(sessionID), userList)

	// Don't close the connection, instead loop endlessly.
	for {
		// Read from our messageChan.
		event, open := <-sub.Channel

		if !open {
			// If our messageChan was closed, this means that the client has
			// disconnected.
			break
		}

		sendEvent(w, f, msg, event.Type, event.GroupID, event.Data)
	}

	log.Infof(msg, r.URL.Path)
}

func sendEvent(
	w http.ResponseWriter,
	f http.Flusher,
	msg string,
	eventType events.EventType,
	groupID events.GroupID,
	payload interface{},
) {
	data, err := json.Marshal(payload)
	if err != nil {
		log.Errorf(msg, err)
	}

	// Write to the ResponseWriter, `w`.
	_, err = fmt.Fprintf(w, "event: %s\ndata: %s\n\n", eventType, data)
	if err != nil {
		log.Errorf(msg, err)
	}

	log.Infof("[sse] sent event: type=%v group=%v", eventType, groupID)

	// Flush the response. This is only possible if
	// the response supports streaming.
	f.Flush()
}
