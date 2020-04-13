package playerctrl

import (
	"context"
	"errors"
	"time"

	"github.com/antonbaumann/spotify-jukebox/player"

	"github.com/antonbaumann/spotify-jukebox/db"
	"github.com/antonbaumann/spotify-jukebox/events"
	"github.com/antonbaumann/spotify-jukebox/sse"
	log "github.com/sirupsen/logrus"
	"github.com/zmb3/spotify"
)

var (
	ErrEventPayloadMalformed = errors.New("event payload malformed")
)

type Controller struct {
	sessionCollection db.SessionCollection
	songCollection    db.SongCollection
	userCollection    db.UserCollection
	playerCollection  db.PlayerCollection

	authenticator spotify.Authenticator

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
	playerCollection db.PlayerCollection,
	authenticator spotify.Authenticator,
) *Controller {
	controller := &Controller{
		sessionCollection: sessionCollection,
		userCollection:    userCollection,
		songCollection:    songCollection,
		playerCollection:  playerCollection,
		authenticator:     authenticator,
		eventBus:          eventBus,
		timers:            make(map[string]*time.Timer),
	}
	return controller
}

func (ctrl *Controller) Start() error {
	sessionIDs, err := ctrl.sessionCollection.ListSessionIDs(context.TODO())
	if err != nil {
		return err
	}

	// initialize timer for every known sessionID
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
	playPause := ctrl.eventBus.Subscribe(
		[]events.EventType{PlayPauseEvent},
		[]events.GroupID{events.GroupIDAny},
	)

	for {
		select {
		case ev := <-playPause.Channel:
			ctrl.handlePlayPause(ev.Type, ev.GroupID, ev.Data)
		}
	}
}

func (ctrl *Controller) setTimer(sessionID string, duration time.Duration, f func()) {
	t, ok := ctrl.timers[sessionID]
	if !ok {
		ctrl.timers[sessionID] = time.AfterFunc(duration, f)
	} else {
		t.Reset(duration)
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

	ctrl.eventBus.Publish(sse.PlaylistChange, events.GroupID(sessionID), songList[1:])

	// fetch next song after song has ended
	ctrl.setTimer(
		sessionID,
		time.Duration(nextSong.Duration)*time.Millisecond,
		func() { ctrl.getNextSong(sessionID) },
	)

	// update session player
	newPlayer := &player.Player{
		CurrentSong:  nextSong,
		SongProgress: 0,
		SongStart:    time.Now(),
		Paused:       false,
	}
	if err := ctrl.playerCollection.SetPlayer(ctx, sessionID, newPlayer); err != nil {
		log.Errorf("%v: %v", msg, err)
	}

	ctrl.notifyClients(
		sessionID,
		ctrl.setPlayerStateAction(
			newPlayer.CurrentSong.ID,
			0,
			false,
		),
	)
}

// synchronizes all connected users with admin player state
func (ctrl *Controller) notifyClients(sessionID string, action notifyAction) {
	msg := "[playerctrl] notify clients"
	ctx := context.Background()

	clients, err := ctrl.userCollection.GetSpotifyClients(ctx, sessionID)
	if err != nil {
		log.Errorf("%v: %v", msg, err)
	}

	for _, client := range clients {
		spotifyClient := ctrl.authenticator.NewClient(client.AuthToken)
		action(spotifyClient)
	}
}
