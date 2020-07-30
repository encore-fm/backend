package playerctrl

import (
	"context"
	"errors"
	"time"

	"github.com/antonbaumann/spotify-jukebox/user"
	log "github.com/sirupsen/logrus"
	"github.com/zmb3/spotify"
)

// maps spotify clients to timers
// timer fires when controller should reattempt to notify client
var notifyTimers = make(map[string]*time.Timer)

const (
	// exponential BackOff with min of (100*2^k)ms and max of 5000ms with a max number of attempts of 50
	minBackOff      = time.Duration(50) * time.Millisecond // will be multiplied by 2^k
	maxBackOff      = time.Duration(2500) * time.Millisecond
	maxAttempts     = 10
	TooManyRequests = 429
)

func retry(operation func() error, clientID string) {
	retryWithAttempts(operation, clientID, 0)
}

// if max attempts are not exceeded, retries the given operation with an exponential backoff retry time
// until the operation succeeds
func retryWithAttempts(operation func() error, clientID string, attempts int) {
	msg := "playerctrl"

	multiplier := time.Duration(2 << attempts)
	backOff := minBackOff * multiplier
	if backOff > maxBackOff || backOff <= 0 {
		backOff = maxBackOff
	}
	if attempts >= maxAttempts {
		log.Warnf("%v: max retry attempts reached. aborting.", msg)
		return
	}
	err := operation()
	if err != nil {
		// too many requests
		spotifyErr := errors.Unwrap(err)
		if e, ok := spotifyErr.(spotify.Error); ok {
			if e.Status == TooManyRequests {
				log.Errorf("spotify rate limit exceeded.")
				return
			}
		}
		log.Warnf("%v, retrying in %v, attempts: %v", err, backOff, attempts)
		// set the timer for the next attempt
		newTimer := time.AfterFunc(backOff, func() { retryWithAttempts(operation, clientID, attempts+1) })
		t, ok := notifyTimers[clientID]
		if !ok {
			notifyTimers[clientID] = newTimer
		} else {
			// if timer was already set, stop and overwrite
			t.Stop()
			notifyTimers[clientID] = newTimer
		}
	}
}

// initializes a user's spotify client and applies the specified notifyAction
// reattempts the action with exponential backoff time at failure (e.g. due to no device being active or other spotify error)
func (ctrl *Controller) notifyClients(clients []*user.SpotifyClient, action notifyAction) {
	for _, client := range clients {
		spotifyClient := ctrl.authenticator.NewClient(client.AuthToken)
		operation := func() error {
			// ensures that user has an active player before executing an action.
			activatePlayer(spotifyClient)
			return action(spotifyClient)
		}
		// keep retrying the api request if it fails
		retry(operation, client.ID)
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

// finds and activates a client's playback device if no active devices are found
func activatePlayer(client spotify.Client) {
	msg := "[playerctrl] activate player"

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
