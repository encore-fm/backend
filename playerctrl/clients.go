package playerctrl

import (
	"context"
	"fmt"
	"github.com/antonbaumann/spotify-jukebox/user"
	log "github.com/sirupsen/logrus"
	"github.com/zmb3/spotify"
	"math"
	"time"
)

// maps spotify clients to timers
// timer fires when controller should reattempt to notify client
var notifyTimers = make(map[string]*time.Timer)

const (
	minBackOff  = time.Duration(50) * time.Millisecond // will be multiplied by 2^k
	maxBackOff  = time.Duration(5000) * time.Millisecond
	maxAttempts = 50
)

func retry(operation func() error, clientID string) {
	retryWithAttempts(operation, clientID, 0)
}

func retryWithAttempts(operation func() error, clientID string, attempts int) {
	msg := "playerctrl"
	// exponential BackOff with min of (100*2^k)ms and max of 5000ms
	minBackOff := minBackOff * (2 << attempts)
	backOff := time.Duration(math.Min(float64(minBackOff), float64(maxBackOff)))
	if attempts >= maxAttempts {
		log.Warnf("%v: max retry attempts reached. aborting.", msg)
		return
	}
	err := operation()
	if err != nil {
		log.Warnf("%v, %v", err, fmt.Sprintf("retrying in %v", backOff))
		newTimer := time.AfterFunc(backOff, func() { retryWithAttempts(operation, clientID, attempts+1) })
		t, ok := notifyTimers[clientID]
		if !ok {
			notifyTimers[clientID] = newTimer
		} else {
			t.Stop()
			notifyTimers[clientID] = newTimer
		}
	}
}

// initializes a user's spotify client and applies the specified notifyAction
// reattempts after a fixed duration at failure (e.g. spotify/network error or no active devices found)
func (ctrl *Controller) notifyClients(clients []*user.SpotifyClient, action notifyAction) {
	for _, client := range clients {
		spotifyClient := ctrl.authenticator.NewClient(client.AuthToken)
		operation := func() error {
			// ensures that user has an active player before executing an action.
			activatePlayer(spotifyClient)
			return action(spotifyClient)
		}
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
