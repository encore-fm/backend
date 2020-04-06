package handlers

import (
	"testing"

	"github.com/antonbaumann/spotify-jukebox/db"
	"github.com/antonbaumann/spotify-jukebox/player"
	"github.com/antonbaumann/spotify-jukebox/spotifycl"
	"github.com/antonbaumann/spotify-jukebox/sse"
	"github.com/stretchr/testify/assert"
	"github.com/zmb3/spotify"
)

func TestNew(t *testing.T) {
	auth := spotify.NewAuthenticator("http://123.de")
	cli := &spotifycl.SpotifyClient{}
	userCol := db.UserCollection(nil)
	sessCol := db.SessionCollection(nil)
	broker := &sse.Broker{}
	ctrl := &player.Controller{}

	expected := &handler{
		spotifyAuthenticator: auth,
		Spotify:              cli,
		UserCollection:       userCol,
		SessionCollection:    sessCol,
		Broker:               broker,
		PlayerCtrl:           ctrl,
	}

	result := New(userCol, sessCol, auth, cli, broker, ctrl)

	assert.Equal(t, expected, result)
}
