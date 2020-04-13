package playerctrl

import (
	"context"
	"errors"
	"github.com/antonbaumann/spotify-jukebox/player"
	"time"

	"github.com/antonbaumann/spotify-jukebox/db"
	"github.com/antonbaumann/spotify-jukebox/events"
	"github.com/antonbaumann/spotify-jukebox/sse"
	log "github.com/sirupsen/logrus"
	"github.com/zmb3/spotify"
)

// todo: don't send events from buggy Spotify SDK player
const DelayCompensation = time.Millisecond * 300

var (
	ErrEventPayloadMalformed = errors.New("event payload malformed")
)

// define state change events
const (
	ControllerStateChangedEvent events.EventType = "controller_player_state_changed"
	AdminStateChangedEvent      events.EventType = "admin_player_state_changed"
)

type StateChangedPayload struct {
	SongID   string `json:"current_track"`
	Duration int    `json:"duration"`
	Position int    `json:"position"`
	Paused   bool   `json:"paused"`
}

// define register session events
const RegisterSessionEvent events.EventType = "register_session"

type RegisterSessionPayload struct {
	SessionID string `json:"session_id"`
}

type Controller struct {
	sessionCollection db.SessionCollection
	songCollection    db.SongCollection
	userCollection    db.UserCollection

	authenticator spotify.Authenticator

	// if new session is created
	// sessionID must be passed to this channel
	Clients chan string

	eventBus events.EventBus

	// maps sessions to timers
	// timer fires when current song ended and new song must be fetched from db
	timers map[string]*time.Timer
}

func NewController(
	eventBus events.EventBus,
	sessionCollection db.SessionCollection,
	songCollection db.SongCollection,
	userCollection db.UserCollection,
	authenticator spotify.Authenticator,
) *Controller {
	controller := &Controller{
		sessionCollection: sessionCollection,
		userCollection:    userCollection,
		songCollection:    songCollection,
		authenticator:     authenticator,
		eventBus:          eventBus,
		Clients:           make(chan string),
		timers:            make(map[string]*time.Timer),
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
	sub := ctrl.eventBus.Subscribe(
		[]events.EventType{RegisterSessionEvent},
		[]events.GroupID{events.GroupIDAny},
	)
	for {
		select {
		case ev := <-sub.Channel:
			payload, ok := ev.Data.(RegisterSessionPayload)
			if !ok {
				log.Errorf("%v: %v", msg, ErrEventPayloadMalformed)
				continue
			}
			log.Infof("%v: id={%v}", msg, payload.SessionID)
			ctrl.setTimer(payload.SessionID, 0, func() { ctrl.getNextSong(payload.SessionID) })
		}
	}
}

func (ctrl *Controller) eventLoop() {

	adminSub := ctrl.eventBus.Subscribe(
		[]events.EventType{AdminStateChangedEvent},
		[]events.GroupID{events.GroupIDAny},
	)

	controllerSub := ctrl.eventBus.Subscribe(
		[]events.EventType{ControllerStateChangedEvent},
		[]events.GroupID{events.GroupIDAny},
	)

	for {
		select {
		case event := <-adminSub.Channel:
			log.Infof("[playerctrl] AdminStateChangeEvent: %v", event.GroupID)
			payload, ok := event.Data.(StateChangedPayload)
			if !ok {
				log.Error("[playerctrl] event payload malformed")
			}
			ctrl.notifyClients(string(event.GroupID), payload, false)

		case ev := <-controllerSub.Channel:
			log.Infof("[playerctrl] ControllerStateChangeEvent: %v", ev.GroupID)
			payload, ok := ev.Data.(StateChangedPayload)
			if !ok {
				log.Error("[playerctrl] event payload malformed")
			}
			ctrl.notifyClients(string(ev.GroupID), payload, true)
		}
	}
}

func (ctrl *Controller) setTimer(sessionID string, duration time.Duration, f func()) {
	t, ok := ctrl.timers[sessionID]
	if !ok {
		ctrl.timers[sessionID] = time.AfterFunc(duration-DelayCompensation, f)
	} else {
		t.Reset(duration - DelayCompensation)
	}
}

func (ctrl *Controller) stopTimer(sessionID string) {
	t, ok := ctrl.timers[sessionID]
	if ok {
		if !t.Stop() {
			<-t.C
		}
		delete(ctrl.timers, sessionID)
	}
}

// gets next song from db and deletes it
// sends event
func (ctrl *Controller) getNextSong(sessionID string) {
	msg := "[playerctrl] get next song from db"
	ctx := context.Background()
	songList, err := ctrl.songCollection.ListSongs(ctx, sessionID)
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
		// todo: wait for songAdded
		log.Infof("%v: %v", msg, "songlist empty - waiting for 1000ms")
		ctrl.setTimer(
			sessionID,
			time.Duration(1000)*time.Millisecond,
			func() { ctrl.getNextSong(sessionID) },
		)
		return
	}

	nextSong := songList[0]

	// remove nextSong from db
	if err := ctrl.songCollection.RemoveSong(ctx, sessionID, nextSong.ID); err != nil {
		log.Errorf("%v: %v", msg, err)
	}

	payload := StateChangedPayload{
		SongID:   nextSong.ID,
		Duration: nextSong.Duration,
		Position: 0,
		Paused:   false,
	}

	ctrl.eventBus.Publish(ControllerStateChangedEvent, events.GroupID(sessionID), payload)
	ctrl.eventBus.Publish(sse.PlaylistChange, events.GroupID(sessionID), songList[1:])

	// fetch next song after song has ended
	ctrl.setTimer(
		sessionID,
		time.Duration(nextSong.Duration)*time.Millisecond,
		func() { ctrl.getNextSong(sessionID) },
	)

	// update session player
	newPlayer := player.Player{
		CurrentSong:  nextSong,
		SongProgress: 0,
		SongStart:    time.Now(),
		Paused:       false,
	}
	if err := ctrl.sessionCollection.SetPlayer(ctx, sessionID, &newPlayer); err != nil {
		log.Errorf("%v: %v", msg, err)
	}
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
		spotifyClient := ctrl.authenticator.NewClient(client.AuthToken)
		opt := &spotify.PlayOptions{
			URIs:       []spotify.URI{TrackURI(stateChange.SongID)},
			PositionMs: stateChange.Position,
		}

		if !stateChange.Paused {
			// usually we dont want to overwrite the state at the
			// admin's "real spotify player" since this function mostly
			// gets called if the admin changes his player state
			if !client.IsAdmin || notifyAdmin {
				if err := spotifyClient.PlayOpt(opt); err != nil {
					log.Errorf("%v: %v", msg, err)
				}
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
