package playerctrl

import (
	"context"
	"errors"
	"fmt"
	"github.com/antonbaumann/spotify-jukebox/user"
	"time"

	"github.com/antonbaumann/spotify-jukebox/song"
	"github.com/antonbaumann/spotify-jukebox/sse"

	"github.com/antonbaumann/spotify-jukebox/db"
	"github.com/antonbaumann/spotify-jukebox/events"
	"github.com/antonbaumann/spotify-jukebox/player"
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
	sessionTimers map[string]*time.Timer

	// maps spotify clients to timers
	clientTimers map[string]*time.Timer
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
		sessionTimers:     make(map[string]*time.Timer),
		clientTimers:      make(map[string]*time.Timer),
	}
	return controller
}

func (ctrl *Controller) Start() error {
	msg := "[playerctrl] start"
	ctx := context.Background()
	sessionIDs, err := ctrl.sessionCollection.ListSessionIDs(ctx)
	if err != nil {
		return err
	}

	// initialize timer for every known sessionID
	for _, sessionID := range sessionIDs {
		playerState, err := ctrl.playerCollection.GetPlayer(ctx, sessionID)
		if err != nil {
			log.Errorf(msg, err)
		}
		if playerState == nil || playerState.IsEmpty() {
			ctrl.setSessionTimer(sessionID, 0, func() { ctrl.getNextSong(sessionID) })
		} else {
			timerDuration := time.Duration(playerState.CurrentSong.Duration)*time.Millisecond - playerState.Progress()
			if timerDuration < 0 {
				timerDuration = 0
			}
			ctrl.setSessionTimer(
				sessionID,
				timerDuration,
				func() { ctrl.getNextSong(sessionID) },
			)
			ctrl.notifyClientsBySessionID(
				sessionID,
				ctrl.setPlayerStateAction(playerState.CurrentSong.ID, playerState.Progress(), playerState.Paused),
			)
		}
	}

	go ctrl.eventLoop()

	return nil
}

func (ctrl *Controller) eventLoop() {
	songAdded := ctrl.eventBus.Subscribe([]events.EventType{SongAdded}, []events.GroupID{events.GroupIDAny})
	playPause := ctrl.eventBus.Subscribe([]events.EventType{PlayPauseEvent}, []events.GroupID{events.GroupIDAny})
	skip := ctrl.eventBus.Subscribe([]events.EventType{SkipEvent}, []events.GroupID{events.GroupIDAny})
	seek := ctrl.eventBus.Subscribe([]events.EventType{SeekEvent}, []events.GroupID{events.GroupIDAny})
	setSynchronized := ctrl.eventBus.Subscribe([]events.EventType{SetSynchronizedEvent}, []events.GroupID{events.GroupIDAny})
	sseConnection := ctrl.eventBus.Subscribe([]events.EventType{SSEConnectionEvent}, []events.GroupID{events.GroupIDAny})
	reset := ctrl.eventBus.Subscribe([]events.EventType{ResetEvent}, []events.GroupID{events.GroupIDAny})

	for {
		select {
		case ev := <-songAdded.Channel:
			ctrl.handleSongAdded(ev)

		case ev := <-playPause.Channel:
			ctrl.handlePlayPause(ev)

		case ev := <-skip.Channel:
			ctrl.handleSkip(ev)

		case ev := <-seek.Channel:
			ctrl.handleSeek(ev)

		case ev := <-setSynchronized.Channel:
			ctrl.handleSetSynchronized(ev)

		case ev := <-sseConnection.Channel:
			ctrl.handleSSEConnection(ev)

		case ev := <-reset.Channel:
			ctrl.handleReset(ev)
		}
	}
}

func (ctrl *Controller) setSessionTimer(sessionID string, duration time.Duration, f func()) {
	ctrl.setTimer(sessionID, duration, f, true)
}

func (ctrl *Controller) setClientTimer(clientID string, duration time.Duration, f func()) {
	ctrl.setTimer(clientID, duration, f, false)
}

func (ctrl *Controller) stopSessionTimer(sessionID string) {
	ctrl.stopTimer(sessionID, true)
}

func (ctrl *Controller) stopClientTimer(clientID string) {
	ctrl.stopTimer(clientID, false)
}

func (ctrl *Controller) setTimer(timerID string, duration time.Duration, f func(), isSessionTimer bool) {
	var t *time.Timer
	var ok bool
	if isSessionTimer {
		t, ok = ctrl.sessionTimers[timerID]
		if !ok {
			ctrl.sessionTimers[timerID] = time.AfterFunc(duration, f)
		} else {
			t.Stop()
			ctrl.sessionTimers[timerID] = time.AfterFunc(duration, f)
		}
	} else {
		t, ok = ctrl.clientTimers[timerID]
		if !ok {
			ctrl.clientTimers[timerID] = time.AfterFunc(duration, f)
		} else {
			t.Stop()
			ctrl.clientTimers[timerID] = time.AfterFunc(duration, f)
		}
	}
}

func (ctrl *Controller) stopTimer(timerID string, isSessionTimer bool) {
	var t *time.Timer
	var ok bool
	if isSessionTimer {
		t, ok = ctrl.sessionTimers[timerID]
		if ok {
			if !t.Stop() {
				<-t.C
			}
			delete(ctrl.sessionTimers, timerID)
		}
	} else {
		t, ok = ctrl.clientTimers[timerID]
		if ok {
			if !t.Stop() {
				<-t.C
			}
			delete(ctrl.clientTimers, timerID)
		}
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
		ctrl.setSessionTimer(
			sessionID,
			time.Duration(500)*time.Millisecond,
			func() { ctrl.getNextSong(sessionID) },
		)
		return
	}
	if len(songList) == 0 {
		// if songList is empty
		// reset player, log error and wait for songAdded
		err = ctrl.playerCollection.SetPlayer(ctx, sessionID, player.New())
		if err != nil {
			log.Errorf("%v: %v", msg, err)
		}
		// explicitly publish a skip event when playlist is empty, or else last song (in player) cannot get skipped
		ctrl.notifyClientsBySessionID(sessionID, ctrl.playerSkipAction())
		log.Warnf("%v: %v", msg, "songlist empty")
		return
	}

	nextSong := songList[0]

	// remove nextSong from db
	if err := ctrl.songCollection.RemoveSong(ctx, sessionID, nextSong.ID); err != nil {
		log.Errorf("%v: %v", msg, err)
	}

	ctrl.eventBus.Publish(sse.PlaylistChange, events.GroupID(sessionID), songList[1:])

	// fetch next song after song has ended
	ctrl.setSessionTimer(
		sessionID,
		time.Duration(nextSong.Duration)*time.Millisecond,
		func() { ctrl.getNextSong(sessionID) },
	)

	// update session player
	newPlayer := &player.Player{
		CurrentSong: nextSong,
		SongStart:   time.Now(),
		PauseStart:  time.Now(),
		Paused:      false,
	}
	if err := ctrl.playerCollection.SetPlayer(ctx, sessionID, newPlayer); err != nil {
		log.Errorf("%v: %v", msg, err)
	}

	// send out a player state change event
	ctrl.notifyPlayerStateChange(sessionID)

	ctrl.notifyClientsBySessionID(
		sessionID,
		ctrl.setPlayerStateAction(
			newPlayer.CurrentSong.ID,
			0,
			false,
		),
	)
}

// finds and activates a client's playback device if no active devices are found
func initializeClient(client spotify.Client) {
	msg := "[playerctrl] initialize client"

	devices, err := client.PlayerDevices()
	if err != nil {
		log.Errorf("%v: %v", msg, err)
	}
	if len(devices) == 0 {
		return
	}
	for _, device := range devices {
		// Active device found. No action needed.
		if device.Active {
			return
		}
	}
	// else, activate the first device in the list
	err = client.TransferPlayback(devices[0].ID, false)
	if err != nil {
		log.Errorf("%v: %v", msg, err)
		return
	}
}

// initializes a user's spotify client and applies the specified notifyAction
// reattempts after a fixed duration at failure (e.g. spotify/network error or no active devices found)
func (ctrl *Controller) notifyClients(clients []*user.SpotifyClient, action notifyAction) {
	msg := "[playerctrl] notify clients"

	for _, client := range clients {
		spotifyClient := ctrl.authenticator.NewClient(client.AuthToken)
		initializeClient(spotifyClient)
		err := action(spotifyClient)
		if err != nil {
			retry := time.Duration(3000) * time.Millisecond
			log.Errorf("%v: %v, %v", msg, err, fmt.Sprintf("retrying in %v", retry))
			ctrl.setTimer(
				client.ID,
				retry,
				func() { ctrl.notifyClients([]*user.SpotifyClient{client}, action) },
				false,
			)
		}
	}
}

// synchronizes the specified user with admin player state
func (ctrl *Controller) notifyClientByUserID(userID string, action notifyAction) {
	msg := "[playerctrl] notify client by user id"
	ctx := context.Background()

	// get user's client
	client, err := ctrl.userCollection.GetSpotifyClient(ctx, userID)
	if err != nil {
		log.Errorf("%v: %v", msg, err)
		return
	}
	ctrl.notifyClients([]*user.SpotifyClient{client}, action)
}

// synchronizes all connected users with admin player state
func (ctrl *Controller) notifyClientsBySessionID(sessionID string, action notifyAction) {
	msg := "[playerctrl] notify clients by session id"
	ctx := context.Background()

	clients, err := ctrl.userCollection.GetSyncedSpotifyClients(ctx, sessionID)
	if err != nil {
		log.Errorf("%v: %v", msg, err)
	}
	ctrl.notifyClients(clients, action)
}

// sends out a player state change event with relevant data about the current player state
func (ctrl *Controller) notifyPlayerStateChange(sessionID string) {
	msg := "[playerctrl] notify player state change"
	ctx := context.Background()

	var currentSong *song.Model
	var isPlaying bool
	var progress int64

	playr, err := ctrl.playerCollection.GetPlayer(ctx, sessionID)
	if err != nil {
		log.Errorf("%v: %v", msg, err)
		return
	}

	if playr != nil {
		currentSong = playr.CurrentSong
		isPlaying = !playr.Paused
		progress = playr.Progress().Milliseconds()
	}

	payload := &sse.PlayerStateChangePayload{
		CurrentSong: currentSong,
		IsPlaying:   isPlaying,
		ProgressMs:  progress,
		Timestamp:   time.Now(),
	}

	ctrl.eventBus.Publish(
		sse.PlayerStateChange,
		events.GroupID(sessionID),
		payload,
	)
}
