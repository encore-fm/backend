package handlers

import (
	"testing"

	"github.com/encore-fm/backend/db"
	"github.com/encore-fm/backend/events"
	"github.com/encore-fm/backend/spotifycl"
	"github.com/stretchr/testify/assert"
	"github.com/zmb3/spotify"
)

func TestNew(t *testing.T) {
	eventBus := events.NewEventBus()
	auth := spotify.NewAuthenticator("http://123.de")
	cli := &spotifycl.SpotifyClient{}
	userCol := db.UserCollection(nil)
	sessCol := db.SessionCollection(nil)
	songCol := db.SongCollection(nil)
	playerCol := db.PlayerCollection(nil)

	expected := &handler{
		eventBus:             eventBus,
		spotifyAuthenticator: auth,
		Spotify:              cli,
		UserCollection:       userCol,
		SessionCollection:    sessCol,
		SongCollection:       songCol,
	}

	result := New(eventBus, userCol, sessCol, songCol, playerCol, auth, cli)

	assert.Equal(t, expected, result)
}
