package playerctrl

import (
	"fmt"
	"time"

	"github.com/zmb3/spotify"
)

type notifyAction = func(client spotify.Client) error

func (ctrl *Controller) setPlayerStateWithOptions(opt *spotify.PlayOptions, paused bool) notifyAction {
	msg := "[playerctrl] set state"
	return func(client spotify.Client) error {
		if !paused {
			if err := client.PlayOpt(opt); err != nil {
				return fmt.Errorf("%v: %v", msg, err)
			}
		} else {
			if err := client.PauseOpt(opt); err != nil {
				return fmt.Errorf("%v: %v", msg, err)
			}
		}
		return nil
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
	return func(client spotify.Client) error {
		if err := client.Next(); err != nil {
			return fmt.Errorf("%v: %v", msg, err)
		}
		return nil
	}
}

func (ctrl *Controller) playerPauseAction() notifyAction {
	msg := "[playerctrl] player pause"
	return func(client spotify.Client) error {
		//// returning errors when attempting to pause causes retry in clients.go to needlessly reattempt to pause client
		//_ = client.Pause()
		//return nil
		if err := client.Pause(); err != nil {
			// exclude 403 error, which causes "restriction violated" bug:
			// if a pause is attempted while player is already paused, spotify throws a "restriction violated" error, which
			// causes the retry in clients.go to keep attempting to pause.
			if err.(spotify.Error).Status == 403 {
				return nil
			}
			return fmt.Errorf("%v: %v", msg, err)
		}
		return nil
	}
}
