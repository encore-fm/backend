package playerctrl

import (
	"context"
	"time"

	"github.com/antonbaumann/spotify-jukebox/events"
	log "github.com/sirupsen/logrus"
)

func (ctrl *Controller) handlePlayPause(
	eventType events.EventType,
	groupID events.GroupID,
	data interface{},
) {
	msg := "[playerctrl] handle play/pause"
	ctx := context.Background()

	sessionID := string(groupID)

	payload, ok := data.(PlayPausePayload)
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

	if !payload.Paused {
		ctrl.setTimer(
			sessionID,
			(time.Duration(p.CurrentSong.Duration)*time.Millisecond)-p.Progress(),
			func() { ctrl.getNextSong(sessionID) },
		)
	} else {
		ctrl.stopTimer(sessionID)
	}

	log.Infof("%v: type={%v} id={%v}", msg, eventType, groupID)
}

func (ctrl *Controller) handleSkip(
	eventType events.EventType,
	groupID events.GroupID,
	data interface{},
) {
	msg := "[playerctrl] handle skip"
	sessionID := string(groupID)
	_, ok := data.(SkipPayload)
	if !ok {
		log.Errorf("%v: %v", msg, ErrEventPayloadMalformed)
		return
	}

	ctrl.getNextSong(sessionID)
	log.Infof("%v: type={%v} id={%v}", msg, eventType, groupID)
}

func (ctrl *Controller) handleSeek(
	eventType events.EventType,
	groupID events.GroupID,
	data interface{},
) {
	ctx := context.Background()
	msg := "[playerctrl] handle seek"
	sessionID := string(groupID)
	payload, ok := data.(SeekPayload)
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

	if !p.Paused {
		songDuration := time.Duration(p.CurrentSong.Duration) * time.Millisecond
		timerDuration := songDuration - payload.Progress
		ctrl.setTimer(
			sessionID,
			timerDuration,
			func() { ctrl.getNextSong(sessionID) },
		)
	}

	log.Infof("%v: type={%v} id={%v}", msg, eventType, groupID)
}
