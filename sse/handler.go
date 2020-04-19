package sse

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/antonbaumann/spotify-jukebox/song"

	"github.com/antonbaumann/spotify-jukebox/events"
	"github.com/gorilla/mux"
	log "github.com/sirupsen/logrus"
)

const (
	PlaylistChange    events.EventType = "sse:playlist_change"
	PlayerStateChange events.EventType = "sse:player_state_change"
)

type PlayerStateChangePayload struct {
	CurrentSong *song.Model   `json:"current_song"`
	IsPlaying   bool          `json:"is_playing"`
	Progress    time.Duration `json:"progress"`
	Timestamp   time.Time	  `json:"timestamp"`
}

type Handler interface {
	ServeHTTP(w http.ResponseWriter, r *http.Request)
}

type handler struct {
	eventBus events.EventBus
}

var _ Handler = (*handler)(nil)

func New(eventBus events.EventBus) Handler {
	return &handler{eventBus: eventBus}
}

// This Broker method handles and HTTP request at the "/users/{session_id}" URL.
//
func (h *handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	msg := "[sse] serve http: %v"

	vars := mux.Vars(r)
	sessionID := vars["session_id"]

	// Make sure that the writer supports flushing.
	//
	f, ok := w.(http.Flusher)
	if !ok {
		http.Error(w, "Streaming unsupported!", http.StatusInternalServerError)
		return
	}

	// subscribe to playlist changes
	sub := h.eventBus.Subscribe(
		[]events.EventType{PlaylistChange, PlayerStateChange},
		[]events.GroupID{events.GroupID(sessionID)},
	)

	// Listen to the closing of the http connection
	go func() {
		<-ctx.Done()
		// Remove this client from the map of attached clients
		// when `EventHandler` exits.

		h.eventBus.Unsubscribe(sub)

		log.Info("[sse] HTTP connection just closed")
	}()

	// Set the headers related to event streaming.
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	w.Header().Set("Transfer-Encoding", "chunked")

	// Don't close the connection, instead loop endlessly.
	for {
		// Read from our messageChan.
		event, open := <-sub.Channel

		if !open {
			// If our messageChan was closed, this means that the client has
			// disconnected.
			break
		}

		data, err := json.Marshal(event.Data)
		if err != nil {
			log.Errorf(msg, err)
		}

		// Write to the ResponseWriter, `w`.
		_, err = fmt.Fprintf(w, "event: %s\ndata: %s\n\n", event.Type, data)
		if err != nil {
			log.Errorf(msg, err)
		}

		log.Infof("[sse] sent event: type=%v group=%v", event.Type, event.GroupID)

		// Flush the response. This is only possible if
		// the response supports streaming.
		f.Flush()
	}

	log.Infof(msg, r.URL.Path)
}
