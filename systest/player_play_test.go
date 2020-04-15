// +build !ci

package systest

import (
	"context"
	"errors"
	"net/http"
	"testing"
	"time"

	"github.com/antonbaumann/spotify-jukebox/player"
	"github.com/antonbaumann/spotify-jukebox/session"
	"github.com/stretchr/testify/assert"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func setPaused() error {
	filter := bson.D{
		{"_id", TestSessionID},
	}

	update := bson.D{
		{
			"$set",
			bson.D{
				{"player.paused", true},
				{"player.song_paused", time.Now()},
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

func getPlayer() (*player.Player, error) {
	filter := bson.D{
		{"_id", TestSessionID},
	}
	projection := bson.D{
		{"_id", 0},
		{"player", 1},
	}

	var sess session.Session
	err := sessionCollection.FindOne(
		context.TODO(),
		filter,
		options.FindOne().SetProjection(projection),
	).Decode(&sess)
	if err != nil {
		return nil, err
	}

	return sess.Player, nil
}

// pauses
// waits few seconds
// player progress should not have changed
func TestPlayerPlay(t *testing.T) {
	dropDB()
	setupDB()

	_, err := resetPlayerController(TestSessionID)
	assert.NoError(t, err)

	pauseDuration := 3 * time.Second

	time.Sleep(2 * time.Second)

	err = setPaused()
	assert.NoError(t, err)

	playerOld, err := getPlayer()
	assert.NoError(t, err)

	time.Sleep(pauseDuration)

	resp, err := PlayerPlay(TestAdminUsername, TestAdminSecret, TestSessionID)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	// short pause because PlayerPlay is async
	time.Sleep(200 * time.Millisecond)
	playerNew, err := getPlayer()
	assert.NoError(t, err)

	assert.WithinDuration(t, testNow.Add(playerOld.Progress()), testNow.Add(playerNew.Progress()), time.Millisecond * 300)
}
