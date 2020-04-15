// +build !ci

package systest

import (
	"context"
	"errors"
	"net/http"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"go.mongodb.org/mongo-driver/bson"
)

func setPlaying() error {
	filter := bson.D{
		{"_id", TestSessionID},
	}

	update := bson.D{
		{
			"$set",
			bson.D{
				{"player.paused", false},
			},
		},
	}

	result, err := sessionCollection.UpdateOne(context.TODO(), filter, update)
	if err != nil {
		return err
	}
	if result.MatchedCount == 0 {
		return errors.New("failed to update player state")
	}
	return nil
}

func TestPlayerPause(t *testing.T) {
	dropDB()
	setupDB()

	_, err := resetPlayerController(TestSessionID)
	assert.NoError(t, err)

	// sleep until controller fetches song
	time.Sleep(2 * time.Second)

	err = setPlaying()
	assert.NoError(t, err)

	resp, err := PlayerPause(TestAdminUsername, TestAdminSecret, TestSessionID)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	// short pause because PlayerPlay is async
	time.Sleep(200 * time.Millisecond)
	p, err := getPlayer()
	assert.NoError(t, err)
	assert.WithinDuration(t, testNow.Add(90 * time.Second), testNow.Add(p.Progress()), 1 * time.Second)
}
