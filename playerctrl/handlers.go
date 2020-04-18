package playerctrl

import (
	"context"
	"github.com/antonbaumann/spotify-jukebox/config"
	"github.com/antonbaumann/spotify-jukebox/player"
	"time"

	"github.com/antonbaumann/spotify-jukebox/events"
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

	// todo: implement SetPaused as findAndUpdate
	p, err := ctrl.playerCollection.GetPlayer(ctx, sessionID)
	if err != nil {
		log.Errorf("%v: %v", msg, err)
		return
	}

	ctrl.notifyClients(sessionID,
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

	delta := p.Progress() - payload.Progress
	if err := ctrl.playerCollection.IncrementProgress(ctx, sessionID, delta); err != nil {
		log.Errorf("%v: %v", msg, err)
		return
	}

	ctrl.notifyClients(sessionID,
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
