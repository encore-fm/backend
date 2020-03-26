package handlers

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"reflect"
	"strings"
	"testing"

	"github.com/antonbaumann/spotify-jukebox/db"
	"github.com/antonbaumann/spotify-jukebox/db/mocks"
	"github.com/antonbaumann/spotify-jukebox/user"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// test if Login behaves correctly
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

// test if Login behaves correctly
// - credentials are correct
// - admin does not exist in database
func TestHandler_Login2(t *testing.T) {
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
