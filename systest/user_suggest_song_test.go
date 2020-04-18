// +build !ci

package systest

import (
	"context"
	"encoding/json"
	"net/http"
	"testing"

	"github.com/antonbaumann/spotify-jukebox/handlers"
	"github.com/antonbaumann/spotify-jukebox/session"
	"github.com/antonbaumann/spotify-jukebox/song"
	"github.com/antonbaumann/spotify-jukebox/util"
	"github.com/stretchr/testify/assert"
	"go.mongodb.org/mongo-driver/bson"
)

func Test_UserSuggestSong_GoodID(t *testing.T) {
	username := TestAdminUsername
	secret := TestAdminSecret
	sessionID := TestSessionID
	songID := NotRickRollSongID

	// get song list count before request
	var foundSession *session.Session
	err := sessionCollection.FindOne(context.Background(), bson.D{{"_id", sessionID}}).Decode(&foundSession)
	assert.NoError(t, err)
	count := len(foundSession.SongList)

	resp, err := UserSuggestSong(username, secret, sessionID, songID)
	assert.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusOK, resp.StatusCode)

	var response song.Model
	err = json.NewDecoder(resp.Body).Decode(&response)
	assert.NoError(t, err)

	assert.Equal(t, songID, response.ID)
	// song score should be 1 after being suggested
	assert.Equal(t, 1, response.Score)
	assert.Equal(t, 1, len(response.Upvoters))
	// make sure user is in upvoters
	index := util.Find(
		len(response.Upvoters),
		func(i int) bool {
			return response.Upvoters[i] == username
		},
	)
	assert.NotEqual(t, -1, index)
	// make sure song is written to db
	err = sessionCollection.FindOne(context.Background(), bson.D{{"_id", sessionID}}).Decode(&foundSession)
	assert.NoError(t, err)
	newCount := len(foundSession.SongList)
	assert.Equal(t, count+1, newCount)
}

func Test_UserSuggestSong_ExistingSong(t *testing.T) {
	username := TestUserName
	secret := TestUserSecret
	sessionID := TestSessionID
	songID := SkiFoanID

	// suggest song twice, since it gets consumed by the player controller the first time
	_, _ = UserSuggestSong(username, secret, sessionID, songID)
	resp, err := UserSuggestSong(username, secret, sessionID, songID)
	assert.NoError(t, err)
	defer resp.Body.Close()

	// 409 expected when song exists
	assert.Equal(t, http.StatusConflict, resp.StatusCode)

	// deserialize response body and assert expected results
	var response handlers.FrontendError

	err = json.NewDecoder(resp.Body).Decode(&response)
	assert.NoError(t, err)
	assert.Equal(t, handlers.SongConflictError, response)
}

func Test_UserSuggestSong_BadID(t *testing.T) {
	username := TestAdminUsername
	secret := TestAdminSecret
	sessionID := TestSessionID
	songID := ""

	// get song list count before request
	var foundSession *session.Session
	err := sessionCollection.FindOne(context.Background(), bson.D{{"_id", sessionID}}).Decode(&foundSession)
	assert.NoError(t, err)
	count := len(foundSession.SongList)

	resp, err := UserSuggestSong(username, secret, sessionID, songID)
	assert.NoError(t, err)
	defer resp.Body.Close()

	// make sure response code is not OK
	assert.NotEqual(t, http.StatusOK, resp.StatusCode)

	// make sure no songs were added to db
	err = sessionCollection.FindOne(context.Background(), bson.D{{"_id", sessionID}}).Decode(&foundSession)
	assert.NoError(t, err)
	newCount := len(foundSession.SongList)
	assert.Equal(t, count, newCount)
}

// this function implicitly tests the user auth function associated with the user handlers
func Test_UserSuggestSong_BadUser(t *testing.T) {
	username := "baumanto"
	secret := ""
	sessionID := TestSessionID
	songID := NotRickRollSongID

	resp, err := UserSuggestSong(username, secret, sessionID, songID)
	assert.NoError(t, err)
	defer resp.Body.Close()

	assert.NotEqual(t, http.StatusOK, resp.StatusCode)
}
