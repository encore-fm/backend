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
	"github.com/antonbaumann/spotify-jukebox/song"
	"github.com/antonbaumann/spotify-jukebox/sse"
	"github.com/gorilla/mux"
	"github.com/stretchr/testify/assert"
)

// - song exists in db
func TestHandler_RemoveSong(t *testing.T) {
	sessionID := "session_id"
	songID := "song_id"
	username := "username"

	// set up userCollection mock
	var sessionCollection db.SessionCollection
	sessionCollection = &mocks.SessionCollection{}

	sessionCollection.(*mocks.SessionCollection).
		On("RemoveSong", context.TODO(), sessionID, songID).
		Return(
			nil,
		)

	sessionCollection.(*mocks.SessionCollection).
		On("ListSongs", context.TODO(), sessionID).
		Return(
			[]*song.Model{},
			nil,
		)

	// take element from chanel to avoid blocking
	ch := make(chan sse.Event)
	go func() {
		<-ch
	}()

	// create handler with mock collections
	handler := &handler{
		SessionCollection: sessionCollection,
		Broker:            &sse.Broker{Notifier: ch},
	}
	adminHandler := AdminHandler(handler)

	// set up http request
	req, err := http.NewRequest(
		"DELETE",
		fmt.Sprintf("/users/username/removeSong/%v", songID),
		nil,
	)
	if err != nil {
		t.Fatal(err)
	}
	req = mux.SetURLVars(req, map[string]string{
		"username": username,
		"song_id":  songID,
	})
	req.Header.Set("Session", sessionID)
	rr := httptest.NewRecorder()

	// call handler func
	adminHandler.RemoveSong(rr, req)

	// Check the status code is what we expect
	if status := rr.Code; status != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusOK)
	}

	// Check the response body is what we expect
	var result []song.Model
	err = json.NewDecoder(rr.Body).Decode(&result)
	assert.Nil(t, err)

	expected := make([]song.Model, 0)
	assert.Equal(t, expected, result)
}

func TestHandler_RemoveSong_NoSessionWithID(t *testing.T) {
	sessionID := "session_id"
	songID := "song_id"
	username := "username"

	// set up sessionCollection mock
	var sessionCollection db.SessionCollection
	sessionCollection = &mocks.SessionCollection{}

	sessionCollection.(*mocks.SessionCollection).
		On("RemoveSong", context.TODO(), sessionID, songID).
		Return(
			db.ErrNoSessionWithID,
		)

	// create handler with mock collections
	handler := &handler{
		SessionCollection: sessionCollection,
	}
	adminHandler := AdminHandler(handler)

	// set up http request
	req, err := http.NewRequest(
		"DELETE",
		fmt.Sprintf("/users/username/removeSong/%v", songID),
		nil,
	)
	if err != nil {
		t.Fatal(err)
	}
	req = mux.SetURLVars(req, map[string]string{
		"username": username,
		"song_id":  songID,
	})
	req.Header.Set("Session", sessionID)
	rr := httptest.NewRecorder()

	// call handler func
	adminHandler.RemoveSong(rr, req)

	// Check the status code is what we expect
	if status := rr.Code; status != http.StatusInternalServerError {
		t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusInternalServerError)
	}

	// Check the response body is what we expect
	assert.Equal(t, fmt.Sprintf("%v\n", db.ErrNoSessionWithID.Error()), rr.Body.String())
}
