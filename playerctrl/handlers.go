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

	if err := ctrl.playerCollection.SetPaused(ctx, sessionID, payload.Paused); err != nil {
		log.Errorf("%v: %v", msg, err)
	}

	// todo: implement SetPaused as findUpdate
	p, err := ctrl.playerCollection.GetPlayer(ctx, sessionID)
	if err != nil {
		log.Errorf("%v: %v", msg, err)
	}

	ctrl.notifyClients(sessionID, ctrl.setPausedAction(payload.Paused))

	if !payload.Paused {
		ctrl.setTimer(
			sessionID,
			(time.Duration(p.CurrentSong.Duration)*time.Millisecond)-p.SongProgress,
			func() { ctrl.getNextSong(sessionID) },
		)
	}

	log.Infof("%v: type={%v} id={%v}", msg, eventType, groupID)
}
