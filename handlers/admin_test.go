package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"reflect"
	"strings"
	"testing"

	"github.com/antonbaumann/spotify-jukebox/db"
	"github.com/antonbaumann/spotify-jukebox/db/mocks"
	"github.com/antonbaumann/spotify-jukebox/song"
	"github.com/antonbaumann/spotify-jukebox/user"
	"github.com/gorilla/mux"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// - credentials are correct
// - admin exists in database
func TestHandler_Login(t *testing.T) {
	// set up userCollection mock
	var userCollection db.UserCollection
	userCollection = &mocks.UserCollection{}

	userInfo := &user.Model{
		Username: "admin",
		Secret:   "secret",
		IsAdmin:  true,
		Score:    0,
	}

	userCollection.(*mocks.UserCollection).
		On("GetUser", "admin").
		Return(
			userInfo,
			nil,
		)

	// create handler with mock collections
	handler := &handler{
		spotifyActivated: false,
		UserCollection:   userCollection,
		SongCollection:   nil,
	}
	adminHandler := AdminHandler(handler)

	// set up http request
	req, err := http.NewRequest(
		"GET",
		"/admin/login",
		strings.NewReader(`{"username": "admin", "password": "password"}`),
	)
	if err != nil {
		t.Fatal(err)
	}
	rr := httptest.NewRecorder()

	// call handler func
	adminHandler.Login(rr, req)

	// Check the status code is what we expect
	if status := rr.Code; status != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusOK)
	}

	// Check the response body is what we expect
	var result *user.Model
	err = json.NewDecoder(rr.Body).Decode(&result)
	assert.Nil(t, err)

	if !reflect.DeepEqual(result, userInfo) {
		t.Errorf("handler returned unexpected body: got %v want %v", result, userInfo)
	}
}

// - credentials are correct
// - admin does not exist in database
func TestHandler_Login_UserNotInDB(t *testing.T) {
	// set up userCollection mock
	var userCollection db.UserCollection
	userCollection = &mocks.UserCollection{}

	userCollection.(*mocks.UserCollection).
		On("GetUser", "admin").
		Return(
			nil,
			nil,
		)

	userCollection.(*mocks.UserCollection).
		On("AddUser", mock.MatchedBy(func(u *user.Model) bool {
			return u.IsAdmin && u.Username == "admin" && u.Score == 0
		})).
		Return(
			nil,
		)

	// create handler with mock collections
	handler := &handler{
		spotifyActivated: false,
		UserCollection:   userCollection,
		SongCollection:   nil,
	}
	adminHandler := AdminHandler(handler)

	// set up http request
	req, err := http.NewRequest(
		"GET",
		"/admin/login",
		strings.NewReader(`{"username": "admin", "password": "password"}`),
	)
	if err != nil {
		t.Fatal(err)
	}
	rr := httptest.NewRecorder()

	// call handler func
	adminHandler.Login(rr, req)

	// Check the status code is what we expect
	if status := rr.Code; status != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusOK)
	}

	// Check the response body is what we expect
	var result *user.Model
	err = json.NewDecoder(rr.Body).Decode(&result)
	assert.Nil(t, err)

	result.Secret = ""

	expected := &user.Model{
		Username: "admin",
		Secret:   "",
		IsAdmin:  true,
		Score:    0,
	}

	if !reflect.DeepEqual(result, expected) {
		t.Errorf("handler returned unexpected body: got %v want %v", result, expected)
	}
}

// - credentials not json
func TestHandler_Login_JsonCorrupted(t *testing.T) {
	// create handler with mock collections
	handler := &handler{
		spotifyActivated: false,
		UserCollection:   nil,
		SongCollection:   nil,
	}
	adminHandler := AdminHandler(handler)

	// set up http request
	req, err := http.NewRequest(
		"GET",
		"/admin/login",
		strings.NewReader(`{"username": "admin", "password": "password"`),
	)
	if err != nil {
		t.Fatal(err)
	}
	rr := httptest.NewRecorder()

	// call handler func
	adminHandler.Login(rr, req)

	// Check the status code is what we expect
	if status := rr.Code; status != http.StatusBadRequest {
		t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusBadRequest)
	}

	expected := "unexpected EOF\n"
	if rr.Body.String() != expected {
		t.Errorf("handler returned unexpected body: got %v want %v", rr.Body.String(), expected)
	}
}

// - wrong credentials
func TestHandler_Login_WrongCredentials(t *testing.T) {
	// create handler with mock collections
	handler := &handler{
		spotifyActivated: false,
		UserCollection:   nil,
		SongCollection:   nil,
	}
	adminHandler := AdminHandler(handler)

	// set up http request
	req, err := http.NewRequest(
		"GET",
		"/admin/login",
		strings.NewReader(`{"username": "user", "password": "12345"}`),
	)
	if err != nil {
		t.Fatal(err)
	}
	rr := httptest.NewRecorder()

	// call handler func
	adminHandler.Login(rr, req)

	// Check the status code is what we expect
	if status := rr.Code; status != http.StatusForbidden {
		t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusForbidden)
	}

	expected := "username and password do not match\n"
	if rr.Body.String() != expected {
		t.Errorf("handler returned unexpected body: got %v want %v", rr.Body.String(), expected)
	}
}

// - song exists in db
func TestHandler_RemoveSong(t *testing.T) {
	// set up userCollection mock
	var songCollection db.SongCollection
	songCollection = &mocks.SongCollection{}

	songCollection.(*mocks.SongCollection).
		On("RemoveSong", "id").
		Return(
			nil,
		)

	songCollection.(*mocks.SongCollection).
		On("ListSongs").
		Return(
			[]*song.Model{},
			nil,
		)

	// create handler with mock collections
	handler := &handler{
		SongCollection: songCollection,
	}
	adminHandler := AdminHandler(handler)

	// set up http request
	req, err := http.NewRequest(
		"GET",
		"/users/username/removeSong/id",
		nil,
	)
	if err != nil {
		t.Fatal(err)
	}
	req = mux.SetURLVars(req, map[string]string{
		"username": "username",
		"song_id": "id",
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
		On("RemoveSong", "id").
		Return(
			ErrSongNotInCollection,
		)

	// create handler with mock collections
	handler := &handler{
		SongCollection: songCollection,
	}
	adminHandler := AdminHandler(handler)

	// set up http request
	req, err := http.NewRequest(
		"GET",
		"/users/username/removeSong/id",
		nil,
	)
	if err != nil {
		t.Fatal(err)
	}
	req = mux.SetURLVars(req, map[string]string{
		"username": "username",
		"song_id": "id",
	})
	rr := httptest.NewRecorder()

	// call handler func
	adminHandler.RemoveSong(rr, req)

	// Check the status code is what we expect
	if status := rr.Code; status != http.StatusInternalServerError {
		t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusInternalServerError)
	}

	// Check the response body is what we expect
	assert.Equal(t, fmt.Sprintf("%v\n", ErrSongNotInCollection.Error()), rr.Body.String())
}