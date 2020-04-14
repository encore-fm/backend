package systest

import (
	"context"
	"errors"
	"testing"

	"github.com/antonbaumann/spotify-jukebox/player"
	"github.com/antonbaumann/spotify-jukebox/session"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func setPaused(paused bool) error {
	filter := bson.D{
		{"_id", TestSessionID},
	}

	update := bson.D{
		{
			"$set",
			bson.D{
				{"player.paused", paused},
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

// todo wait until seek endpoint exist and reset controller state after each test
func TestPlayerPlay(t *testing.T) {
	//dropDB()
	//setupDB()
	//
	//err := setPaused(true)
	//assert.NoError(t, err)
	//
	//resp, err := PlayerPlay(TestAdminUsername, TestAdminSecret, TestSessionID)
	//assert.NoError(t, err)
	//assert.Equal(t, http.StatusOK, resp.StatusCode)
	//
	//// short pause because PlayerPlay is async
	//time.Sleep(200 * time.Millisecond)
	//p, err := getPlayer()
	//assert.NoError(t, err)
	//
	//assert.WithinDuration(t, testNow.Add(30 * time.Second), testNow.Add(p.Progress()), time.Millisecond * 300)
}
