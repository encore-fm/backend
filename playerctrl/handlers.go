package playerctrl

import (
	"context"
	"time"

	"github.com/encore-fm/backend/config"
	"github.com/encore-fm/backend/player"
	"github.com/encore-fm/backend/sse"

	"github.com/encore-fm/backend/events"
	log "github.com/sirupsen/logrus"
)

func (ctrl *Controller) handleSongAdded(ev events.Event) {
	msg := "[playerctrl] handle song added"
	ctx := context.Background()

	sessionID := string(ev.GroupID)

	playr, err := ctrl.playerCollection.GetPlayer(ctx, sessionID)
	if err != nil {
		log.Errorf("%v: %v", msg, err)
		return
	}
	// timer only needs to be set when the player is empty
	if playr != nil && !playr.IsEmpty() {
		return
	}

	ctrl.setTimer(sessionID, 0, func() { ctrl.getNextSong(sessionID) })
}

func (ctrl *Controller) handlePlayPause(ev events.Event) {
	msg := "[playerctrl] handle play/pause"
	ctx := context.Background()

	sessionID := string(ev.GroupID)

	// todo: implement SetPaused as findAndUpdate
	p, err := ctrl.playerCollection.GetPlayer(ctx, sessionID)
	if err != nil {
		log.Errorf("%v: %v", msg, err)
		return
	}

	if p.IsEmpty() {
		log.Warnf("%v: no song in player", msg)
		return
	}

	payload, ok := ev.Data.(PlayPausePayload)
	if !ok {
		log.Errorf("%v: %v", msg, ErrEventPayloadMalformed)
		return
	}

	if payload.Paused {
		if err := ctrl.playerCollection.SetPaused(ctx, sessionID); err != nil {
			log.Errorf("%v: %v", msg, err)
			return
		}
	} else {
		if err := ctrl.playerCollection.SetPlaying(ctx, sessionID); err != nil {
			log.Errorf("%v: %v", msg, err)
			return
		}
	}

	ctrl.notifyClientsBySessionID(sessionID,
		ctrl.setPlayerStateAction(
			p.CurrentSong.ID,
			p.Progress(),
			payload.Paused,
		),
	)
	// send out a player state change event
	ctrl.notifyPlayerStateChange(sessionID)

	if !payload.Paused {
		ctrl.setTimer(
			sessionID,
			(time.Duration(p.CurrentSong.Duration)*time.Millisecond)-p.Progress(),
			func() { ctrl.getNextSong(sessionID) },
		)
	} else {
		ctrl.stopTimer(sessionID)
	}

	log.Infof("%v: type={%v} id={%v}", msg, ev.Type, ev.GroupID)
}

func (ctrl *Controller) handleSkip(ev events.Event) {
	msg := "[playerctrl] handle skip"
	sessionID := string(ev.GroupID)
	_, ok := ev.Data.(SkipPayload)
	if !ok {
		log.Errorf("%v: %v", msg, ErrEventPayloadMalformed)
		return
	}

	ctrl.getNextSong(sessionID)

	// send out a player state change event
	ctrl.notifyPlayerStateChange(sessionID)

	log.Infof("%v: type={%v} id={%v}", msg, ev.Type, ev.GroupID)
}

func (ctrl *Controller) handleSeek(ev events.Event) {
	ctx := context.Background()
	msg := "[playerctrl] handle seek"
	sessionID := string(ev.GroupID)
	payload, ok := ev.Data.(SeekPayload)
	if !ok {
		log.Errorf("%v: %v", msg, ErrEventPayloadMalformed)
		return
	}

	// todo: find a way to make atomic
	// todo: currently only one user per session can manipulate player

	p, err := ctrl.playerCollection.GetPlayer(ctx, sessionID)
	if err != nil {
		log.Errorf("%v: %v", msg, err)
		return
	}

	// if no song is in player, no further action is needed
	if p.IsEmpty() {
		log.Warnf("%v: no song in player", msg)
		return
	}

	delta := p.Progress() - payload.Progress
	if err := ctrl.playerCollection.IncrementProgress(ctx, sessionID, delta); err != nil {
		log.Errorf("%v: %v", msg, err)
		return
	}

	ctrl.notifyClientsBySessionID(sessionID,
		ctrl.setPlayerStateAction(
			p.CurrentSong.ID,
			payload.Progress,
			p.Paused,
		),
	)

	// send out a player state change event
	ctrl.notifyPlayerStateChange(sessionID)

	if !p.Paused {
		songDuration := time.Duration(p.CurrentSong.Duration) * time.Millisecond
		timerDuration := songDuration - payload.Progress
		ctrl.setTimer(
			sessionID,
			timerDuration,
			func() { ctrl.getNextSong(sessionID) },
		)
	}

	log.Infof("%v: type={%v} id={%v}", msg, ev.Type, ev.GroupID)
}

func (ctrl *Controller) handleSetSynchronized(ev events.Event) {
	msg := "[playerctrl] handle set synchronized"
	ctx := context.Background()
	sessionID := string(ev.GroupID)

	payload, ok := ev.Data.(SetSynchronizedPayload)
	if !ok {
		log.Errorf("%v: %v", msg, ErrEventPayloadMalformed)
		return
	}
	userID := payload.UserID
	synchronized := payload.Synchronized

	var err error
	if synchronized {
		err = ctrl.synchronizeUser(sessionID, userID)
	} else {
		err = ctrl.desynchronizeUser(userID)
	}
	if err != nil {
		log.Errorf("%v: %v", msg, err)
		return
	}
	// notify sse that user change sync status
	ctrl.eventBus.Publish(
		sse.UserSynchronizedChange,
		events.GroupID(sessionID),
		sse.UserSynchronizedChangePayload{Synchronized: synchronized, UserID: userID},
	)

	// notify sse that user list changed
	userList, err := ctrl.userCollection.ListUsers(ctx, sessionID)
	if err != nil {
		log.Errorf("%v: %v", msg, err)
	}
	if userList != nil {
		ctrl.eventBus.Publish(
			sse.UserListChange,
			events.GroupID(sessionID),
			userList,
		)
	}

	log.Infof("%v: type={%v} id={%v}", msg, ev.Type, ev.GroupID)
}

// handles new and removed sse connections. The connectionEstablished flag specified whether it's a new incoming
// connection or the removal of an old connection
func (ctrl *Controller) handleSSEConnection(ev events.Event) {
	msg := "[playerctrl] handle sse connections"
	ctx := context.Background()
	sessionID := string(ev.GroupID)
	payload, ok := ev.Data.(SSEConnectionPayload)
	if !ok {
		log.Errorf("%v: %v", msg, ErrEventPayloadMalformed)
		return
	}
	userID := payload.UserID
	connectionEstablished := payload.ConnectionEstablished

	usr, err := ctrl.userCollection.GetUserByID(ctx, userID)
	if err != nil {
		log.Errorf("%v: %v", msg, ErrEventPayloadMalformed)
		return
	}

	// user doesn't want to be synced/desynced automatically -> do nothing.
	if !usr.AutoSync {
		return
	}
	// otherwise, sync/desync user
	synchronize := connectionEstablished // connection established -> sync, otherwise desync.
	if synchronize {
		err = ctrl.synchronizeUser(sessionID, userID)
	} else {
		err = ctrl.desynchronizeUser(userID)
	}
	if err != nil {
		log.Errorf("%v: %v", msg, err)
		return
	}
	// notify sse that user change sync status
	ctrl.eventBus.Publish(
		sse.UserSynchronizedChange,
		events.GroupID(sessionID),
		sse.UserSynchronizedChangePayload{Synchronized: synchronize, UserID: userID},
	)

	// notify sse that user list changed
	userList, err := ctrl.userCollection.ListUsers(ctx, sessionID)
	if err != nil {
		log.Errorf("%v: %v", msg, err)
	}
	if userList != nil {
		ctrl.eventBus.Publish(
			sse.UserListChange,
			events.GroupID(sessionID),
			userList,
		)
	}

	log.Infof("%v: type={%v} id={%v}", msg, ev.Type, ev.GroupID)
}

func (ctrl *Controller) synchronizeUser(sessionID, userID string) error {
	ctx := context.Background()

	err := ctrl.userCollection.SetSynchronized(ctx, userID, true)
	if err != nil {
		return err
	}

	// get player to extract current playing information
	playr, err := ctrl.playerCollection.GetPlayer(ctx, sessionID)
	if err != nil {
		return err
	}

	// if no songs in session, pause the client
	if playr.IsEmpty() {
		ctrl.notifyClientByUserID(userID, ctrl.playerPauseAction())
		return nil
	}

	// get the user's client up to speed...
	ctrl.notifyClientByUserID(
		userID,
		ctrl.setPlayerStateAction(
			playr.CurrentSong.ID,
			playr.Progress(),
			playr.Paused,
		),
	)
	return nil
}

func (ctrl *Controller) desynchronizeUser(userID string) error {
	ctx := context.Background()

	err := ctrl.userCollection.SetSynchronized(ctx, userID, false)
	if err != nil {
		return err
	}

	// pause the client when the user desynchronizes
	ctrl.notifyClientByUserID(userID, ctrl.playerPauseAction())
	return nil
}

func (ctrl *Controller) handleReset(ev events.Event) {
	msg := "[playerctrl] handle reset"
	ctx := context.Background()
	sessionID := string(ev.GroupID)

	if !config.Conf.Server.Debug {
		log.Errorf("%v: debug event sent but running in production mode", msg)
		return
	}
	payload, ok := ev.Data.(ResetPayload)
	if !ok {
		log.Errorf("%v: reset event: %v", msg, ErrEventPayloadMalformed)
		return
	}

	// setup test player todo: getPlayer in player_play_test.go still returns nil
	err := ctrl.playerCollection.SetPlayer(ctx, sessionID, player.New())
	if err != nil {
		log.Errorf("%v: %v", msg, err)
		return
	}

	log.Infof("%v: id={%v}", msg, payload.SessionID)
	ctrl.setTimer(payload.SessionID, 0, func() { ctrl.getNextSong(payload.SessionID) })
}
