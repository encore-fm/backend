package handlers

import (
	"testing"

	"github.com/antonbaumann/spotify-jukebox/db"
	"github.com/antonbaumann/spotify-jukebox/events"
	"github.com/antonbaumann/spotify-jukebox/spotifycl"
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

	expected := &handler{
		eventBus:             eventBus,
		spotifyAuthenticator: auth,
		Spotify:              cli,
		UserCollection:       userCol,
		SessionCollection:    sessCol,
		SongCollection:       songCol,
	}

	result := New(eventBus, userCol, sessCol, songCol, auth, cli)

	assert.Equal(t, expected, result)
}
