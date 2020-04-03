package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/antonbaumann/spotify-jukebox/db"
	"github.com/antonbaumann/spotify-jukebox/db/mocks"
	"github.com/antonbaumann/spotify-jukebox/session"
	"github.com/antonbaumann/spotify-jukebox/song"
	"github.com/antonbaumann/spotify-jukebox/sse"
	"github.com/antonbaumann/spotify-jukebox/user"
	"github.com/gorilla/mux"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/zmb3/spotify"
)

// test successful join request
func TestHandler_Join(t *testing.T) {
	sessionID := "session_id"
	username := "username"

	// set up sessionCollection mock
	var sessionCollection db.SessionCollection
	sessionCollection = &mocks.SessionCollection{}

	// set up userCollection mock
	var userCollection db.UserCollection
	userCollection = &mocks.UserCollection{}

	// GetSessionByID successful
	sessionCollection.(*mocks.SessionCollection).
		On("GetSessionByID", context.TODO(), sessionID).
		Return(
			&session.Session{ID: sessionID, SongList: make([]*song.Model, 0)},
			nil,
		)

	// no error if correct user is added
	userCollection.(*mocks.UserCollection).
		On("AddUser", context.TODO(), mock.MatchedBy(func(u *user.Model) bool {
			return u.Username == username &&
				u.SessionID == sessionID &&
				u.ID == user.GenerateUserID(username, sessionID)
		})).
		Return(nil)

	// create handler with mock collections
	handler := &handler{
		UserCollection:       userCollection,
		SessionCollection:    sessionCollection,
		spotifyAuthenticator: spotify.NewAuthenticator("http://123.de"),
	}
	userHandler := UserHandler(handler)

	// set up http request
	req, err := http.NewRequest(
		"DELETE",
		fmt.Sprintf("/users/%v/join/%v", username, sessionID),
		nil,
	)
	if err != nil {
		t.Fatal(err)
	}
	req = mux.SetURLVars(req, map[string]string{
		"username":   username,
		"session_id": sessionID,
	})
	rr := httptest.NewRecorder()

	// call handler func
	userHandler.Join(rr, req)

	// Check the status code is what we expect
	assert.Equal(t, http.StatusOK, rr.Code)

	// decode response body
	var response *struct {
		UserInfo *user.Model `json:"user_info"`
		AuthUrl  string      `json:"auth_url"`
	}
	err = json.NewDecoder(rr.Body).Decode(&response)
	assert.NoError(t, err)

	assert.NotEmpty(t, response.AuthUrl)
	assert.Equal(t, user.GenerateUserID(username, sessionID), response.UserInfo.ID)
	assert.Equal(t, username, response.UserInfo.Username)
	assert.Equal(t, sessionID, response.UserInfo.SessionID)
}

// test successful user list request
func TestHandler_ListUsers(t *testing.T) {
	username := "username"
	sessionID := "session_id"
	userList := make([]*user.ListElement, 0)

	// set up userCollection mock
	var userCollection db.UserCollection
	userCollection = &mocks.UserCollection{}

	// no error
	userCollection.(*mocks.UserCollection).
		On("ListUsers", context.TODO(), sessionID).
		Return(userList, nil)

	// create handler with mock collections
	handler := &handler{
		UserCollection: userCollection,
	}
	userHandler := UserHandler(handler)

	// set up http request
	req, err := http.NewRequest(
		"GET",
		fmt.Sprintf("/users/%v/listUsers", username),
		nil,
	)
	assert.NoError(t, err)

	req = mux.SetURLVars(req, map[string]string{
		"username": username,
	})
	req.Header.Set("Session", sessionID)
	rr := httptest.NewRecorder()

	// call handler func
	userHandler.ListUsers(rr, req)

	// check for success
	assert.Equal(t, http.StatusOK, rr.Code)

	// decode response
	var response []*user.ListElement
	err = json.NewDecoder(rr.Body).Decode(&response)
	assert.NoError(t, err)

	assert.Equal(t, userList, response)
}

// test successful song suggestion
func TestHandler_SuggestSong(t *testing.T) {
	// todo: find a way to mock spotify.Client
}

// test successful song list request
func TestHandler_ListSongs(t *testing.T) {
	username := "username"
	sessionID := "session_id"
	songList := make([]*song.Model, 0)

	// set up sessionCollection mock
	var sessionCollection db.SessionCollection
	sessionCollection = &mocks.SessionCollection{}

	// no error
	sessionCollection.(*mocks.SessionCollection).
		On("ListSongs", context.TODO(), sessionID).
		Return(songList, nil)

	// create handler with mock collections
	handler := &handler{
		SessionCollection: sessionCollection,
	}
	userHandler := UserHandler(handler)

	// set up http request
	req, err := http.NewRequest(
		"GET",
		fmt.Sprintf("/users/%v/listSongs", username),
		nil,
	)
	assert.NoError(t, err)

	req = mux.SetURLVars(req, map[string]string{
		"username": username,
	})
	req.Header.Set("Session", sessionID)
	rr := httptest.NewRecorder()

	// call handler func
	userHandler.ListSongs(rr, req)

	// check for success
	assert.Equal(t, http.StatusOK, rr.Code)

	// decode response
	var response []*song.Model
	err = json.NewDecoder(rr.Body).Decode(&response)
	assert.NoError(t, err)

	assert.Equal(t, songList, response)
}

func TestHandler_Vote(t *testing.T) {
	username := "username"
	sessionID := "session_id"
	songID := "song_id"
	scoreChange := +1
	suggestingUser := "test_user"
	songInitScore := 0

	songInfo := &song.Model{
		ID:          songID,
		SuggestedBy: suggestingUser,
		Score:       songInitScore,
	}

	songList := []*song.Model{songInfo}

	// set up sessionCollection mock
	var sessionCollection db.SessionCollection
	sessionCollection = &mocks.SessionCollection{}

	// set up userCollection mock
	var userCollection db.UserCollection
	userCollection = &mocks.UserCollection{}

	// no error on vote up
	sessionCollection.(*mocks.SessionCollection).
		On("VoteUp", context.TODO(), sessionID, songID, username).
		Return(scoreChange, nil)

	// no error on GetSongByID
	sessionCollection.(*mocks.SessionCollection).
		On("GetSongByID", context.TODO(), sessionID, songID).
		Return(songInfo, nil)

	// no error on ListSongs
	sessionCollection.(*mocks.SessionCollection).
		On("ListSongs", context.TODO(), sessionID).
		Return(songList, nil)

	// no errors incrementing score
	userCollection.(*mocks.UserCollection).
		On("IncrementScore",
			context.TODO(),
			user.GenerateUserID(suggestingUser, sessionID),
			scoreChange,
		).
		Return(nil)

	// take element from chanel to avoid blocking
	ch := make(chan sse.Event)
	go func() {
		<-ch
	}()

	// create handler with mock collections
	handler := &handler{
		SessionCollection: sessionCollection,
		UserCollection:    userCollection,
		Broker:            &sse.Broker{Notifier: ch},
	}
	userHandler := UserHandler(handler)

	// set up http request
	req, err := http.NewRequest(
		"POST",
		fmt.Sprintf("/users/%v/vote/%v/up", username, songID),
		nil,
	)
	assert.NoError(t, err)

	req = mux.SetURLVars(req, map[string]string{
		"username":    username,
		"song_id":     songID,
		"vote_action": "up",
	})
	req.Header.Set("Session", sessionID)
	rr := httptest.NewRecorder()

	// call handler func
	userHandler.Vote(rr, req)

	// check for success
	assert.Equal(t, http.StatusOK, rr.Code)

	// decode response
	var response []*song.Model
	err = json.NewDecoder(rr.Body).Decode(&response)
	assert.NoError(t, err)

	assert.Equal(t, songList, response)
}