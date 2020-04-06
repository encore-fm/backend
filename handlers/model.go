package handlers

import (
	"github.com/antonbaumann/spotify-jukebox/db"
	"github.com/antonbaumann/spotify-jukebox/player"
	"github.com/antonbaumann/spotify-jukebox/spotifycl"
	"github.com/antonbaumann/spotify-jukebox/sse"
	"github.com/zmb3/spotify"
)

type handler struct {
	spotifyAuthenticator spotify.Authenticator
	Spotify              *spotifycl.SpotifyClient
	UserCollection       db.UserCollection
	SessionCollection    db.SessionCollection
	Broker               *sse.Broker
	PlayerCtrl           *player.Controller
}

func New(
	userCollection db.UserCollection,
	sessCollection db.SessionCollection,
	auth spotify.Authenticator,
	client *spotifycl.SpotifyClient,
	broker *sse.Broker,
	playerCtrl *player.Controller,
) *handler {
	return &handler{
		spotifyAuthenticator: auth,
		Spotify:              client,
		UserCollection:       userCollection,
		SessionCollection:    sessCollection,
		Broker:               broker,
		PlayerCtrl:           playerCtrl,
	}
}
