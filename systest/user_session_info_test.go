// +build !ci

package systest

import (
	"encoding/json"
	"net/http"
	"testing"

	"github.com/encore-fm/backend/handlers"
	"github.com/encore-fm/backend/song"
	"github.com/stretchr/testify/assert"
)

func Test_UserGetSessionInfo_InvalidSession(t *testing.T) {
	dropDB()
	setupDB()

	// totally valid sessionID
	resp, err := UserGetSessionInfo("watch?v=dQw4w9WgXcQ")
	assert.NoError(t, err)
	// invalid session -> bad request
	assert.Equal(t, http.StatusBadRequest, resp.StatusCode)

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
