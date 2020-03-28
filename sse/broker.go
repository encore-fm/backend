package sse

import (
	"encoding/json"
	"fmt"
	"net/http"

	log "github.com/sirupsen/logrus"
)

func NewBroker() *Broker {
	return &Broker{
		clients:        make(map[chan Event]bool),
		newClients:     make(chan (chan Event)),
		defunctClients: make(chan (chan Event)),
		Notifier:       make(chan Event),
	}
}

type EventType string

const (
	PlaylistChange EventType = "playlist_change"
)

type Event struct {
	Event EventType
	Data  interface{}
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
	clients map[chan Event]bool

	// Channel into which new clients can be pushed
	//
	newClients chan chan Event

	// Channel into which disconnected clients should be pushed
	//
	defunctClients chan chan Event

	// Channel into which messages are pushed to be broadcast out
	// to attached clients.
	//
	Notifier chan Event
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
				b.clients[s] = true
				log.Info("Added new client")

			case s := <-b.defunctClients:

				// A client has detached and we want to
				// stop sending them Notifier.
				delete(b.clients, s)
				close(s)

				log.Info("Removed client")

			case msg := <-b.Notifier:

				// There is a new message to send.  For each
				// attached client, push the new message
				// into the client's message channel.
				for s := range b.clients {
					s <- msg
				}
				log.Infof("Broadcast message to %d clients", len(b.clients))
			}
		}
	}()
}

// This Broker method handles and HTTP request at the "/events/" URL.
//
func (b *Broker) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	msg := "[broker] serve http: %v"

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

	// Add this client to the map of those that should
	// receive updates
	b.newClients <- messageChan

	// Listen to the closing of the http connection via the CloseNotifier
	go func() {
		<-ctx.Done()
		// Remove this client from the map of attached clients
		// when `EventHandler` exits.
		b.defunctClients <- messageChan
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
