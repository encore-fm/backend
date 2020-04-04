package sse

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/gorilla/mux"
	log "github.com/sirupsen/logrus"
)

func NewBroker() *Broker {
	return &Broker{
		clients:        make(map[string]map[chan Event]bool),
		newClients:     make(chan clientConn),
		defunctClients: make(chan clientConn),
		Notifier:       make(chan Event),
	}
}

type EventType string

const (
	PlaylistChange EventType = "playlist_change"
)

type Event struct {
	GroupID string
	Event   EventType
	Data    interface{}
}

type clientConn struct {
	GroupID      string
	EventChannel chan Event
}

// A single Broker will be created in this program. It is responsible
// for keeping a list of which clients (browsers) are currently attached
// and broadcasting events (Notifier) to those clients.
//
type Broker struct {

	// Create a map of clients, the keys of the map are the channels
	// over which we can push Notifier to attached clients.  (The values
	// are just booleans and are meaningless.)
	//
	clients map[string]map[chan Event]bool

	// Channel into which new clients can be pushed
	//
	newClients chan clientConn

	// Channel into which disconnected clients should be pushed
	//
	defunctClients chan clientConn

	// Channel into which messages are pushed to be broadcast out
	// to attached clients.
	//
	Notifier chan Event
}

func (b *Broker) AddConnection(conn clientConn) {
	group, ok := b.clients[conn.GroupID]
	if ok {
		group[conn.EventChannel] = true
	} else {
		b.clients[conn.GroupID] = make(map[chan Event]bool)
		b.clients[conn.GroupID][conn.EventChannel] = true
	}
}

// This Broker method starts a new goroutine.  It handles
// the addition & removal of clients, as well as the broadcasting
// of Notifier out to clients that are currently attached.
//
func (b *Broker) Start() {

	// Start a goroutine
	//
	go func() {

		// Loop endlessly
		//
		for {

			// Block until we receive from one of the
			// three following channels.
			select {

			case s := <-b.newClients:

				// There is a new client attached and we
				// want to start sending them Notifier.
				b.AddConnection(s)
				log.Infof("[sse] added new client to GroupID: [%v]", s.GroupID)

			case s := <-b.defunctClients:

				// A client has detached and we want to
				// stop sending them Notifier.
				delete(b.clients[s.GroupID], s.EventChannel)
				close(s.EventChannel)

				log.Infof("[sse] removed client from GroupID: [%v]", s.GroupID)

			case msg := <-b.Notifier:

				// There is a new message to send.  For each
				// attached client, push the new message
				// into the client's message channel.
				group, ok := b.clients[msg.GroupID]
				if ok {
					for s := range group {
						s <- msg
					}
				}
				log.Infof(
					"[sse] broadcast message to %d clients in GroupID [%v]",
					len(b.clients[msg.GroupID]),
					msg.GroupID,
				)
			}
		}
	}()
}

// This Broker method handles and HTTP request at the "/users/{username}/events" URL.
//
func (b *Broker) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	msg := "[broker] serve http: %v"

	vars := mux.Vars(r)
	sessionID := vars["session_id"]

	// Make sure that the writer supports flushing.
	//
	f, ok := w.(http.Flusher)
	if !ok {
		http.Error(w, "Streaming unsupported!", http.StatusInternalServerError)
		return
	}

	// Create a new channel, over which the broker can
	// send this client Notifier.
	messageChan := make(chan Event)
	conn := clientConn{
		GroupID:      sessionID,
		EventChannel: messageChan,
	}

	// Add this client to the map of those that should
	// receive updates
	b.newClients <- conn

	// Listen to the closing of the http connection
	go func() {
		<-ctx.Done()
		// Remove this client from the map of attached clients
		// when `EventHandler` exits.
		b.defunctClients <- conn
		log.Info("HTTP connection just closed.")
	}()

	// Set the headers related to event streaming.
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	w.Header().Set("Transfer-Encoding", "chunked")

	// Don't close the connection, instead loop endlessly.
	for {

		// Read from our messageChan.
		event, open := <-messageChan

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
		_, err = fmt.Fprintf(w, "event: %s\ndata: %s\n\n", event.Event, data)
		if err != nil {
			log.Errorf(msg, err)
		}

		// Flush the response. This is only possible if
		// the response supports streaming.
		f.Flush()
	}

	// Done.
	log.Infof(msg, r.URL.Path)
}
