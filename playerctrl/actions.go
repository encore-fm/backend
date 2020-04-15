package playerctrl

import (
	"time"

	log "github.com/sirupsen/logrus"
	"github.com/zmb3/spotify"
)

type notifyAction = func(client spotify.Client)

func (ctrl *Controller) setPlayerStateWithOptions(opt *spotify.PlayOptions, paused bool) notifyAction {
	msg := "[spotifyctrl] set state"
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
