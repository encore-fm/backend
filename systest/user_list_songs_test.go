// +build !ci

package systest

import (
	"context"
	"encoding/json"
	"net/http"
	"testing"

	"github.com/encore-fm/backend/session"
	"github.com/encore-fm/backend/song"
	"github.com/stretchr/testify/assert"
	"go.mongodb.org/mongo-driver/bson"
)

func Test_UserListSongs(t *testing.T) {
	username := TestAdminUsername
	secret := TestAdminSecret
	sessionID := TestSessionID

	// get song list count before request
	var foundSession *session.Session
	err := sessionCollection.FindOne(
		context.Background(),
		bson.D{
			{"_id", sessionID},
		},
	).Decode(&foundSession)
	assert.NoError(t, err)
	count := len(foundSession.SongList)

	resp, err := UserListSongs(username, secret, sessionID)
	assert.NoError(t, err)
	defer resp.Body.Close()

	// make sure response code is OK
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	var response []*song.Model

	err = json.NewDecoder(resp.Body).Decode(&response)
	assert.NoError(t, err)

	// make sure count matches response
	assert.Equal(t, count, len(response))
}
