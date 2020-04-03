package systest

import (
	"encoding/json"
	"fmt"
	"github.com/antonbaumann/spotify-jukebox/handlers"
	"github.com/antonbaumann/spotify-jukebox/song"
	"github.com/antonbaumann/spotify-jukebox/user"
	"github.com/antonbaumann/spotify-jukebox/util"
	"github.com/stretchr/testify/assert"
	"io/ioutil"
	"net/http"
	"os"
	"testing"
)

const (
	BackendBaseUrl    = "http://127.0.0.1:8080"
	ExistingAdminName = "baumanto"
	ExistingSessionID = "1"
	NotRickRollSongID = "4uLU6hMCjMI75M1A2tKUQC"
)

func TestMain(m *testing.M) {
	// todo configure test db
	os.Exit(m.Run())
	// todo whipe test db
}

// Tests adding a new user to an existing session. Expects normal behavior.
func Test_UserJoin_ExistingSession(t *testing.T) {
	username := "jonhue"
	sessionID := ExistingSessionID

	endpointUrl := fmt.Sprintf(BackendBaseUrl+"/users/%v/join/%v", username, sessionID)

	resp, err := http.Post(endpointUrl, "", nil)
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()

	assert.Equal(t, http.StatusOK, resp.StatusCode)

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		t.Fatal(err)
	}

	response := &struct {
		UserInfo *user.Model `json:"user_info"`
		AuthUrl  string      `json:"auth_url"`
	}{}

	err = json.Unmarshal(body, response)
	if err != nil {
		t.Fatal(err)
	}

	assert.Equal(t, username, response.UserInfo.Username)
	assert.Equal(t, ExistingSessionID, response.UserInfo.SessionID)
	assert.Equal(t, 1, response.UserInfo.Score)
	assert.Equal(t, false, response.UserInfo.IsAdmin)
}

func Test_UserJoin_NonExistingSession(t *testing.T) {
	username := "eti"
	sessionID, err := util.GenerateSecret()

	if err != nil {
		t.Fatal(err)
	}

	endpointUrl := fmt.Sprintf(BackendBaseUrl+"/users/%v/join/%v", username, sessionID)

	resp, err := http.Post(endpointUrl, "", nil)
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()

	assert.Equal(t, http.StatusNotFound, resp.StatusCode) // 404 expected when session not found

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		t.Fatal(err)
	}

	response := handlers.FrontendError{}

	err = json.Unmarshal(body, response)
	if err != nil {
		t.Fatal(err)
	}

	assert.Equal(t, handlers.SessionNotFoundError, response) // Assert correct frontend error
}

func Test_UserJoin_ExistingUser(t *testing.T) {
	username := ExistingAdminName
	sessionID := ExistingSessionID

	endpointUrl := fmt.Sprintf(BackendBaseUrl+"/users/%v/join/%v", username, sessionID)

	resp, err := http.Post(endpointUrl, "", nil)
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()

	assert.Equal(t, http.StatusConflict, resp.StatusCode) // 409 expected when username exists

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		t.Fatal(err)
	}

	response := handlers.FrontendError{}

	err = json.Unmarshal(body, response)
	if err != nil {
		t.Fatal(err)
	}

	assert.Equal(t, handlers.UserConflictError, response) // Assert correct frontend error
}

func Test_UserList_OK(t *testing.T) {
	username := ExistingAdminName

	endpointUrl := fmt.Sprintf(BackendBaseUrl+"/users/%v/list", username)

	resp, err := http.Get(endpointUrl)
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()

	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func Test_UserSuggestSong_GoodID(t *testing.T) {
	username := ExistingAdminName
	songID := NotRickRollSongID

	endpointUrl := fmt.Sprintf(BackendBaseUrl+"/users/%v/suggest/%v", username, songID)

	resp, err := http.Post(endpointUrl, "", nil)
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()

	assert.Equal(t, http.StatusOK, resp.StatusCode)

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		t.Fatal(err)
	}

	response := song.Model{}

	err = json.Unmarshal(body, response)
	if err != nil {
		t.Fatal(err)
	}

	assert.Equal(t, songID, response.ID)
	assert.Equal(t, 1, response.Score) // song score should be 1 after being suggested
	assert.Equal(t, 1, len(response.Upvoters))
	assert.True(t, util.Contains(response.Upvoters, username))
}

func Test_UserSuggestSong_BadID(t *testing.T) {
	username := ExistingAdminName
	songID := ""

	endpointUrl := fmt.Sprintf(BackendBaseUrl+"/users/%v/suggest/%v", username, songID)

	resp, err := http.Post(endpointUrl, "", nil)
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()

	assert.NotEqual(t, http.StatusOK, resp.StatusCode)
}

func Test_UserSuggestSong_BadUser(t *testing.T) {
	username := ""
	songID := NotRickRollSongID

	endpointUrl := fmt.Sprintf(BackendBaseUrl+"/users/%v/suggest/%v", username, songID)

	resp, err := http.Post(endpointUrl, "", nil)
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()

	assert.NotEqual(t, http.StatusOK, resp.StatusCode)
}
