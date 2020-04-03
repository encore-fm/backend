package handlers

import (
	"testing"

	"github.com/antonbaumann/spotify-jukebox/db"
	"github.com/antonbaumann/spotify-jukebox/sse"
	"github.com/stretchr/testify/assert"
	"github.com/zmb3/spotify"
)

func TestNew(t *testing.T) {
	auth := spotify.NewAuthenticator("http://123.de")
	cli := auth.NewClient(nil)
	userCol := db.UserCollection(nil)
	sessCol := db.SessionCollection(nil)
	broker := &sse.Broker{}

	expected := &handler{
		spotifyAuthenticator: auth,
		Spotify:              cli,
		UserCollection:       userCol,
		SessionCollection:    sessCol,
		Broker:               broker,
	}

	result := New(userCol, sessCol, auth, cli, broker)

	assert.Equal(t, expected, result)
}
