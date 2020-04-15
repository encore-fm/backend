package handlers

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"
	"time"

	"github.com/antonbaumann/spotify-jukebox/db"
	"github.com/antonbaumann/spotify-jukebox/db/mocks"
	"github.com/antonbaumann/spotify-jukebox/events"
	"github.com/antonbaumann/spotify-jukebox/playerctrl"
	"github.com/antonbaumann/spotify-jukebox/user"
	"github.com/gorilla/mux"
	"github.com/stretchr/testify/assert"
)

func TestHandler_Play(t *testing.T) {
	username := "username"
	sessionID := "sessionID"
	userID := user.GenerateUserID(username, sessionID)

	admin, err := user.NewAdmin(username, sessionID)
	assert.NoError(t, err)

	var userCollection db.UserCollection
	userCollection = &mocks.UserCollection{}

	userCollection.(*mocks.UserCollection).
		On("GetUserByID", context.TODO(), userID).
		Return(
			admin,
			nil,
		)

	eventBus := events.NewEventBus()
	eventBus.Start()
	defer eventBus.Stop()

	handler := &handler{
		eventBus:       eventBus,
		UserCollection: userCollection,
	}

	playerHandler := PlayerHandler(handler)

	// set up http request
	req, err := http.NewRequest(
		"POST",
		fmt.Sprintf("/users/%v/player/play", username),
		nil,
	)
	if err != nil {
		t.Fatal(err)
	}
	req = mux.SetURLVars(req, map[string]string{
		"username": username,
	})
	req.Header.Set("Session", sessionID)
	rr := httptest.NewRecorder()

	sub := handler.eventBus.Subscribe([]events.EventType{playerctrl.PlayPauseEvent}, []events.GroupID{events.GroupID(sessionID)})
	ch := make(chan events.Event)

	go func() {
		for {
			select {
			case ev := <-sub.Channel:
				ch <- ev
				return
			}
		}
	}()

	// call handler func
	playerHandler.Play(rr, req)
	assert.Equal(t, rr.Code, http.StatusOK)

	// wait for event
	ev := <-ch

	assert.Equal(t, string(ev.GroupID), sessionID)
	assert.Equal(t, ev.Type, playerctrl.PlayPauseEvent)

	payload, ok := ev.Data.(playerctrl.PlayPausePayload)
	assert.True(t, ok)
	assert.Equal(t, false, payload.Paused)
}

func TestHandler_Play_NotAdmin(t *testing.T) {
	username := "username"
	sessionID := "sessionID"
	userID := user.GenerateUserID(username, sessionID)

	testUser, err := user.New(username, sessionID)
	assert.NoError(t, err)

	var userCollection db.UserCollection
	userCollection = &mocks.UserCollection{}

	userCollection.(*mocks.UserCollection).
		On("GetUserByID", context.TODO(), userID).
		Return(
			testUser,
			nil,
		)

	eventBus := events.NewEventBus()
	eventBus.Start()
	defer eventBus.Stop()

	handler := &handler{
		eventBus:       eventBus,
		UserCollection: userCollection,
	}

	playerHandler := PlayerHandler(handler)

	// set up http request
	req, err := http.NewRequest(
		"POST",
		fmt.Sprintf("/users/%v/player/play", username),
		nil,
	)
	if err != nil {
		t.Fatal(err)
	}
	req = mux.SetURLVars(req, map[string]string{
		"username": username,
	})
	req.Header.Set("Session", sessionID)
	rr := httptest.NewRecorder()

	// call handler func
	playerHandler.Play(rr, req)
	assert.Equal(t, rr.Code, http.StatusUnauthorized)

	var response FrontendError
	err = json.NewDecoder(rr.Body).Decode(&response)
	assert.NoError(t, err)

	assert.Equal(t, ActionNotAllowedError, response)
}

func TestHandler_Play_NoUserWithID(t *testing.T) {
	username := "username"
	sessionID := "sessionID"
	userID := user.GenerateUserID(username, sessionID)

	var userCollection db.UserCollection
	userCollection = &mocks.UserCollection{}

	userCollection.(*mocks.UserCollection).
		On("GetUserByID", context.TODO(), userID).
		Return(
			nil,
			db.ErrNoUserWithID,
		)

	eventBus := events.NewEventBus()
	eventBus.Start()
	defer eventBus.Stop()

	handler := &handler{
		eventBus:       eventBus,
		UserCollection: userCollection,
	}

	playerHandler := PlayerHandler(handler)

	// set up http request
	req, err := http.NewRequest(
		"POST",
		fmt.Sprintf("/users/%v/player/play", username),
		nil,
	)
	if err != nil {
		t.Fatal(err)
	}
	req = mux.SetURLVars(req, map[string]string{
		"username": username,
	})
	req.Header.Set("Session", sessionID)
	rr := httptest.NewRecorder()

	// call handler func
	playerHandler.Play(rr, req)
	assert.Equal(t, rr.Code, http.StatusUnauthorized)

	var response FrontendError
	err = json.NewDecoder(rr.Body).Decode(&response)
	assert.NoError(t, err)

	assert.Equal(t, RequestNotAuthorizedError, response)
}

func TestHandler_Play_InternalError(t *testing.T) {
	username := "username"
	sessionID := "sessionID"
	userID := user.GenerateUserID(username, sessionID)

	var userCollection db.UserCollection
	userCollection = &mocks.UserCollection{}

	userCollection.(*mocks.UserCollection).
		On("GetUserByID", context.TODO(), userID).
		Return(
			nil,
			errors.New("test"),
		)

	handler := &handler{
		UserCollection: userCollection,
	}

	playerHandler := PlayerHandler(handler)

	// set up http request
	req, err := http.NewRequest(
		"POST",
		fmt.Sprintf("/users/%v/player/play", username),
		nil,
	)
	if err != nil {
		t.Fatal(err)
	}
	req = mux.SetURLVars(req, map[string]string{
		"username": username,
	})
	req.Header.Set("Session", sessionID)
	rr := httptest.NewRecorder()

	// call handler func
	playerHandler.Play(rr, req)
	assert.Equal(t, rr.Code, http.StatusInternalServerError)

	var response FrontendError
	err = json.NewDecoder(rr.Body).Decode(&response)
	assert.NoError(t, err)

	assert.Equal(t, InternalServerError, response)
}

func TestHandler_Pause(t *testing.T) {
	username := "username"
	sessionID := "sessionID"
	userID := user.GenerateUserID(username, sessionID)

	admin, err := user.NewAdmin(username, sessionID)
	assert.NoError(t, err)

	var userCollection db.UserCollection
	userCollection = &mocks.UserCollection{}

	userCollection.(*mocks.UserCollection).
		On("GetUserByID", context.TODO(), userID).
		Return(
			admin,
			nil,
		)

	eventBus := events.NewEventBus()
	eventBus.Start()
	defer eventBus.Stop()

	handler := &handler{
		eventBus:       eventBus,
		UserCollection: userCollection,
	}

	playerHandler := PlayerHandler(handler)

	// set up http request
	req, err := http.NewRequest(
		"POST",
		fmt.Sprintf("/users/%v/player/pause", username),
		nil,
	)
	if err != nil {
		t.Fatal(err)
	}
	req = mux.SetURLVars(req, map[string]string{
		"username": username,
	})
	req.Header.Set("Session", sessionID)
	rr := httptest.NewRecorder()

	sub := handler.eventBus.Subscribe([]events.EventType{playerctrl.PlayPauseEvent}, []events.GroupID{events.GroupID(sessionID)})
	ch := make(chan events.Event)

	go func() {
		for {
			select {
			case ev := <-sub.Channel:
				ch <- ev
				return
			}
		}
	}()

	// call handler func
	playerHandler.Pause(rr, req)
	assert.Equal(t, rr.Code, http.StatusOK)

	// wait for event
	ev := <-ch

	assert.Equal(t, string(ev.GroupID), sessionID)
	assert.Equal(t, ev.Type, playerctrl.PlayPauseEvent)

	payload, ok := ev.Data.(playerctrl.PlayPausePayload)
	assert.True(t, ok)
	assert.Equal(t, true, payload.Paused)
}

func TestHandler_Skip(t *testing.T) {
	username := "username"
	sessionID := "sessionID"
	userID := user.GenerateUserID(username, sessionID)

	admin, err := user.NewAdmin(username, sessionID)
	assert.NoError(t, err)

	var userCollection db.UserCollection
	userCollection = &mocks.UserCollection{}

	userCollection.(*mocks.UserCollection).
		On("GetUserByID", context.TODO(), userID).
		Return(
			admin,
			nil,
		)

	eventBus := events.NewEventBus()
	eventBus.Start()
	defer eventBus.Stop()

	handler := &handler{
		eventBus:       eventBus,
		UserCollection: userCollection,
	}

	playerHandler := PlayerHandler(handler)

	// set up http request
	req, err := http.NewRequest(
		"POST",
		fmt.Sprintf("/users/%v/player/skip", username),
		nil,
	)
	if err != nil {
		t.Fatal(err)
	}
	req = mux.SetURLVars(req, map[string]string{
		"username": username,
	})
	req.Header.Set("Session", sessionID)
	rr := httptest.NewRecorder()

	sub := handler.eventBus.Subscribe([]events.EventType{playerctrl.SkipEvent}, []events.GroupID{events.GroupID(sessionID)})
	ch := make(chan events.Event)

	go func() {
		for {
			select {
			case ev := <-sub.Channel:
				ch <- ev
				return
			}
		}
	}()

	// call handler func
	playerHandler.Skip(rr, req)
	assert.Equal(t, rr.Code, http.StatusOK)

	// wait for event
	ev := <-ch

	assert.Equal(t, string(ev.GroupID), sessionID)
	assert.Equal(t, ev.Type, playerctrl.SkipEvent)

	_, ok := ev.Data.(playerctrl.SkipPayload)
	assert.True(t, ok)
}

func TestHandler_Skip_NotAdmin(t *testing.T) {
	username := "username"
	sessionID := "sessionID"
	userID := user.GenerateUserID(username, sessionID)

	testUser, err := user.New(username, sessionID)
	assert.NoError(t, err)

	var userCollection db.UserCollection
	userCollection = &mocks.UserCollection{}

	userCollection.(*mocks.UserCollection).
		On("GetUserByID", context.TODO(), userID).
		Return(
			testUser,
			nil,
		)

	eventBus := events.NewEventBus()
	eventBus.Start()
	defer eventBus.Stop()

	handler := &handler{
		eventBus:       eventBus,
		UserCollection: userCollection,
	}

	playerHandler := PlayerHandler(handler)

	// set up http request
	req, err := http.NewRequest(
		"POST",
		fmt.Sprintf("/users/%v/player/skip", username),
		nil,
	)
	if err != nil {
		t.Fatal(err)
	}
	req = mux.SetURLVars(req, map[string]string{
		"username": username,
	})
	req.Header.Set("Session", sessionID)
	rr := httptest.NewRecorder()

	// call handler func
	playerHandler.Skip(rr, req)
	assert.Equal(t, rr.Code, http.StatusUnauthorized)

	var response FrontendError
	err = json.NewDecoder(rr.Body).Decode(&response)
	assert.NoError(t, err)

	assert.Equal(t, ActionNotAllowedError, response)
}

func TestHandler_Seek(t *testing.T) {
	username := "username"
	sessionID := "sessionID"
	userID := user.GenerateUserID(username, sessionID)
	progress := time.Second * 30

	admin, err := user.NewAdmin(username, sessionID)
	assert.NoError(t, err)

	var userCollection db.UserCollection
	userCollection = &mocks.UserCollection{}

	userCollection.(*mocks.UserCollection).
		On("GetUserByID", context.TODO(), userID).
		Return(
			admin,
			nil,
		)

	eventBus := events.NewEventBus()
	eventBus.Start()
	defer eventBus.Stop()

	handler := &handler{
		eventBus:       eventBus,
		UserCollection: userCollection,
	}

	playerHandler := PlayerHandler(handler)

	// set up http request
	req, err := http.NewRequest(
		"POST",
		fmt.Sprintf("/users/%v/player/seek/%v", username, strconv.Itoa(int(progress.Milliseconds()))),
		nil,
	)
	if err != nil {
		t.Fatal(err)
	}
	req = mux.SetURLVars(req, map[string]string{
		"username":    username,
		"position_ms": strconv.Itoa(int(progress.Milliseconds())),
	})
	req.Header.Set("Session", sessionID)
	rr := httptest.NewRecorder()

	sub := handler.eventBus.Subscribe([]events.EventType{playerctrl.SeekEvent}, []events.GroupID{events.GroupID(sessionID)})
	ch := make(chan events.Event)

	go func() {
		for {
			select {
			case ev := <-sub.Channel:
				ch <- ev
				return
			}
		}
	}()

	// call handler func
	playerHandler.Seek(rr, req)
	assert.Equal(t, rr.Code, http.StatusOK)

	// wait for event
	ev := <-ch

	assert.EqualValues(t, ev.GroupID, sessionID)
	assert.Equal(t, ev.Type, playerctrl.SeekEvent)

	payload, ok := ev.Data.(playerctrl.SeekPayload)
	assert.True(t, ok)

	assert.Equal(t, progress, payload.Progress)
}

func TestHandler_Seek_NotAdmin(t *testing.T) {
	username := "username"
	sessionID := "sessionID"
	userID := user.GenerateUserID(username, sessionID)
	progress := time.Second * 30

	testUser, err := user.New(username, sessionID)
	assert.NoError(t, err)

	var userCollection db.UserCollection
	userCollection = &mocks.UserCollection{}

	userCollection.(*mocks.UserCollection).
		On("GetUserByID", context.TODO(), userID).
		Return(
			testUser,
			nil,
		)

	eventBus := events.NewEventBus()
	eventBus.Start()
	defer eventBus.Stop()

	handler := &handler{
		eventBus:       eventBus,
		UserCollection: userCollection,
	}

	playerHandler := PlayerHandler(handler)

	// set up http request
	req, err := http.NewRequest(
		"POST",
		fmt.Sprintf("/users/%v/player/seek/%v", username, strconv.Itoa(int(progress.Milliseconds()))),
		nil,
	)
	if err != nil {
		t.Fatal(err)
	}
	req = mux.SetURLVars(req, map[string]string{
		"username": username,
		"position_ms": strconv.Itoa(int(progress.Milliseconds())),
	})
	req.Header.Set("Session", sessionID)
	rr := httptest.NewRecorder()

	// call handler func
	playerHandler.Seek(rr, req)
	assert.Equal(t, rr.Code, http.StatusUnauthorized)

	var response FrontendError
	err = json.NewDecoder(rr.Body).Decode(&response)
	assert.NoError(t, err)

	assert.Equal(t, ActionNotAllowedError, response)
}

func TestHandler_Seek_UrlMalformed(t *testing.T) {
	username := "username"
	sessionID := "sessionID"
	progress := "123malformed"

	eventBus := events.NewEventBus()
	eventBus.Start()
	defer eventBus.Stop()

	handler := &handler{
		eventBus:       eventBus,
	}

	playerHandler := PlayerHandler(handler)

	// set up http request
	req, err := http.NewRequest(
		"POST",
		fmt.Sprintf("/users/%v/player/seek/%v", username, progress),
		nil,
	)
	if err != nil {
		t.Fatal(err)
	}
	req = mux.SetURLVars(req, map[string]string{
		"username": username,
		"position_ms": progress,
	})
	req.Header.Set("Session", sessionID)
	rr := httptest.NewRecorder()

	// call handler func
	playerHandler.Seek(rr, req)
	assert.Equal(t, rr.Code, http.StatusBadRequest)

	var response FrontendError
	err = json.NewDecoder(rr.Body).Decode(&response)
	assert.NoError(t, err)

	assert.Equal(t, RequestUrlMalformedError, response)
}