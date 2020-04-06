package player

import (
	"context"
	"time"

	"github.com/antonbaumann/spotify-jukebox/db"
	"github.com/davecgh/go-spew/spew"
	log "github.com/sirupsen/logrus"
	"github.com/zmb3/spotify"
)

type EventType string

const (
	ControllerStateChangedEvent EventType = "controller_player_state_changed"
	AdminStateChangedEvent      EventType = "admin_player_state_changed"
	UserStateChangedEvent       EventType = "user_player_state_changed"
)

type StateChangedPayload struct {
	SongID   string `json:"current_track"`
	Duration int    `json:"duration"`
	Position int    `json:"position"`
	Paused   bool   `json:"paused"`
}

type Event struct {
	SessionID string
	Type      EventType
	Payload   interface{}
}

type Controller struct {
	sessionCollection db.SessionCollection
	userCollection    db.UserCollection

	authenticator spotify.Authenticator

	// if new session is created
	// sessionID must be passed to this channel
	Clients chan string

	// channel for incoming events
	Events chan Event

	// maps sessions to timers
	// timer fires when current song ended and new song must be fetched from db
	Timers map[string]*time.Timer
}

func NewController(
	sessionCollection db.SessionCollection,
	userCollection db.UserCollection,
	authenticator spotify.Authenticator,
) *Controller {
	controller := &Controller{
		sessionCollection: sessionCollection,
		userCollection:    userCollection,
		authenticator:     authenticator,
		Events:            make(chan Event),
		Clients:           make(chan string),
		Timers:            make(map[string]*time.Timer),
	}
	return controller
}

func (ctrl *Controller) Start() error {
	sessionIDs, err := ctrl.sessionCollection.ListSessionIDs(context.TODO())
	if err != nil {
		return err
	}

	for _, sessionID := range sessionIDs {
		ctrl.setTimer(sessionID, 0, func() { ctrl.getNextSong(sessionID) })
	}

	go ctrl.eventLoop()
	go ctrl.registerSessionLoop()

	return nil
}

func (ctrl *Controller) registerSessionLoop() {
	msg := "[playerctrl] register session"
	for {
		select {
		case sessionID := <-ctrl.Clients:
			log.Infof("%v: id={%v}", msg, sessionID)
			ctrl.setTimer(sessionID, 0, func() { ctrl.getNextSong(sessionID) })
		}
	}
}

func (ctrl *Controller) eventLoop() {
	for {
		select {
		case event := <-ctrl.Events:
			switch event.Type {
			case AdminStateChangedEvent:
				log.Info("[playerctrl] AdminStateChangeEvent")
				payload, ok := event.Payload.(StateChangedPayload)
				if !ok {
					log.Error("[playerctrl] event payload malformed")
				}
				ctrl.notifyClients(event.SessionID, payload, false)

			case ControllerStateChangedEvent:
				log.Info("[playerctrl] ControllerStateChangeEvent")
				payload, ok := event.Payload.(StateChangedPayload)
				if !ok {
					log.Error("[playerctrl] event payload malformed")
				}
				ctrl.notifyClients(event.SessionID, payload, true)

			case UserStateChangedEvent:
				log.Info("[playerctrl] UserStateChangeEvent")
				spew.Dump(event)

			default:
				spew.Dump(event)

			}
		}
	}
}

func (ctrl *Controller) setTimer(sessionID string, duration time.Duration, f func()) {
	t, ok := ctrl.Timers[sessionID]
	if !ok {
		ctrl.Timers[sessionID] = time.AfterFunc(duration, f)
	} else {
		t.Reset(duration)
	}
}

func (ctrl *Controller) stopTimer(sessionID string) {
	t, ok := ctrl.Timers[sessionID]
	if ok {
		if !t.Stop() {
			<-t.C
		}
		delete(ctrl.Timers, sessionID)
	}
}

func (ctrl *Controller) getNextSong(sessionID string) {
	msg := "[playerctrl] get next song from db"
	ctx := context.Background()
	songList, err := ctrl.sessionCollection.ListSongs(ctx, sessionID)
	if err != nil {
		// if error occurs while fetching list
		// log error and try again in 500ms
		log.Errorf("%v: %v", msg, err)
		ctrl.setTimer(
			sessionID,
			time.Duration(500)*time.Millisecond,
			func() { ctrl.getNextSong(sessionID) },
		)
		return
	}
	if len(songList) == 0 {
		// if songList is empty
		// log error and try again in 500ms
		log.Infof("%v: %v", msg, "songlist empty - waiting for 1000ms")
		ctrl.setTimer(
			sessionID,
			time.Duration(1000)*time.Millisecond,
			func() { ctrl.getNextSong(sessionID) },
		)
		return
	}

	nextSong := songList[0]

	payload := StateChangedPayload{
		SongID:   nextSong.ID,
		Duration: nextSong.Duration,
		Position: 0,
		Paused:   false,
	}
	event := Event{
		SessionID: sessionID,
		Type:      ControllerStateChangedEvent,
		Payload:   payload,
	}
	ctrl.Events <- event

	// fetch next song after song has ended
	ctrl.setTimer(
		sessionID,
		time.Duration(nextSong.Duration)*time.Millisecond,
		func() { ctrl.getNextSong(sessionID) },
	)
}

// synchronizes all connected users with admin player state
func (ctrl *Controller) notifyClients(sessionID string, stateChange StateChangedPayload, notifyAdmin bool) {
	msg := "[playerctrl] notify clients"
	ctx := context.Background()

	clients, err := ctrl.userCollection.GetSpotifyClients(ctx, sessionID)
	if err != nil {
		log.Errorf("%v: %v", msg, err)
	}

	for _, client := range clients {
		// don't change admin state
		if client.IsAdmin && !notifyAdmin {
			continue
		}

		spotifyClient := ctrl.authenticator.NewClient(client.AuthToken)
		opt := &spotify.PlayOptions{
			URIs:       []spotify.URI{TrackURI(stateChange.SongID)},
			PositionMs: stateChange.Position,
		}

		if !stateChange.Paused {
			if err := spotifyClient.PlayOpt(opt); err != nil {
				log.Errorf("%v: %v", msg, err)
			}

			delta := time.Millisecond * time.Duration(stateChange.Duration-stateChange.Position)
			ctrl.setTimer(sessionID, delta, func() { ctrl.getNextSong(sessionID) })
		} else {
			if err := spotifyClient.PauseOpt(opt); err != nil {
				log.Errorf("%v: %v", msg, err)
			}
			ctrl.stopTimer(sessionID)
		}
	}
}
