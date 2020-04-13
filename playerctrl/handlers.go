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
		}
	} else {
		if err := ctrl.playerCollection.SetPlaying(ctx, sessionID); err != nil {
			log.Errorf("%v: %v", msg, err)
		}
	}

	// todo: implement SetPaused as findAndUpdate
	p, err := ctrl.playerCollection.GetPlayer(ctx, sessionID)
	if err != nil {
		log.Errorf("%v: %v", msg, err)
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
	}

	log.Infof("%v: type={%v} id={%v}", msg, eventType, groupID)
}
