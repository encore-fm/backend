package playerctrl

import (
	"time"

	log "github.com/sirupsen/logrus"
	"github.com/zmb3/spotify"
)

type notifyAction = func(client spotify.Client)

func (ctrl *Controller) setPlayerStateWithOptions(opt *spotify.PlayOptions, paused bool) notifyAction {
	msg := "[playerctrl] set state"
	return func(client spotify.Client) {
		if !paused {
			if err := client.PlayOpt(opt); err != nil {
				log.Errorf("%v: %v", msg, err)
			}
		} else {
			if err := client.PauseOpt(opt); err != nil {
				log.Errorf("%v: %v", msg, err)
			}
		}
	}
}

func (ctrl *Controller) setPlayerStateAction(songID string, position time.Duration, paused bool) notifyAction {
	opt := &spotify.PlayOptions{
		URIs:       []spotify.URI{TrackURI(songID)},
		PositionMs: int(position.Milliseconds()),
	}
	return ctrl.setPlayerStateWithOptions(opt, paused)
}

// Returns a function that notifies the client to skip to the next song.
// Required when skip request is made on an empty queue
func (ctrl *Controller) playerSkipAction() notifyAction {
	msg := "[playerctrl] player skip"
	return func(client spotify.Client) {
		if err := client.Next(); err != nil {
			log.Errorf("%v: %v", msg, err)
		}
	}
}
