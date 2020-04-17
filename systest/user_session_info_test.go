// +build !ci

package systest

import (
	"encoding/json"
	"github.com/antonbaumann/spotify-jukebox/handlers"
	"github.com/antonbaumann/spotify-jukebox/song"
	"github.com/stretchr/testify/assert"
	"testing"
)

func Test_UserGetSessionInfo_InvalidSession(t *testing.T) {
	dropDB()
	setupDB()

	// empty string for invalid session
	resp, err := UserGetSessionInfo("")
	assert.NoError(t, err)

	var response handlers.FrontendError
	err = json.NewDecoder(resp.Body).Decode(&response)
	assert.NoError(t, err)

	assert.Equal(t, handlers.SessionNotFoundError, response)
}

func Test_UserGetSessionInfo_ValidSession(t *testing.T) {
	dropDB()
	setupDB()

	resp, err := UserGetSessionInfo(TestSessionID)
	assert.NoError(t, err)

	response := &struct {
		AdminName   string      `json:"admin_name"`
		CurrentSong *song.Model `json:"current_song"`
	}{}
	err = json.NewDecoder(resp.Body).Decode(response)
	assert.NoError(t, err)

	assert.Equal(t, TestAdminUsername, response.AdminName)
	assert.Nil(t, response.CurrentSong)
}
