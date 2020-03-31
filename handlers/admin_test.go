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
	// set up userCollection mock
	var songCollection db.SongCollection
	songCollection = &mocks.SongCollection{}

	songCollection.(*mocks.SongCollection).
		On("RemoveSong", context.TODO(), "id").
		Return(
			nil,
		)

	songCollection.(*mocks.SongCollection).
		On("ListSongs", context.TODO()).
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
		SongCollection: songCollection,
		Broker:         &sse.Broker{Notifier: ch},
	}
	adminHandler := AdminHandler(handler)

	// set up http request
	req, err := http.NewRequest(
		"DELETE",
		"/users/username/removeSong/id",
		nil,
	)
	if err != nil {
		t.Fatal(err)
	}
	req = mux.SetURLVars(req, map[string]string{
		"username": "username",
		"song_id":  "id",
	})
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

func TestHandler_RemoveSong_SongNotInDB(t *testing.T) {
	// set up userCollection mock
	var songCollection db.SongCollection
	songCollection = &mocks.SongCollection{}

	songCollection.(*mocks.SongCollection).
		On("RemoveSong", context.TODO(), "id").
		Return(
			db.ErrNoSongWithID,
		)

	// create handler with mock collections
	handler := &handler{
		SongCollection: songCollection,
	}
	adminHandler := AdminHandler(handler)

	// set up http request
	req, err := http.NewRequest(
		"DELETE",
		"/users/username/removeSong/id",
		nil,
	)
	if err != nil {
		t.Fatal(err)
	}
	req = mux.SetURLVars(req, map[string]string{
		"username": "username",
		"song_id":  "id",
	})
	rr := httptest.NewRecorder()

	// call handler func
	adminHandler.RemoveSong(rr, req)

	// Check the status code is what we expect
	if status := rr.Code; status != http.StatusInternalServerError {
		t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusInternalServerError)
	}

	// Check the response body is what we expect
	assert.Equal(t, fmt.Sprintf("%v\n", db.ErrNoSongWithID.Error()), rr.Body.String())
}
