package handlers

import (
	"github.com/antonbaumann/spotify-jukebox/db"
	"github.com/antonbaumann/spotify-jukebox/sse"
	"github.com/zmb3/spotify"
)

type handler struct {
	spotifyAuthenticator spotify.Authenticator
	Spotify              spotify.Client
	spotifyActivated     bool
	UserCollection       db.UserCollection
	SongCollection       db.SongCollection
	SessionCollection    db.SessionCollection
	Broker               *sse.Broker
}

func New(
	userCollection db.UserCollection,
	songCollection db.SongCollection,
	sessCollection db.SessionCollection,
	auth spotify.Authenticator,
	broker *sse.Broker,
) *handler {
	return &handler{
		spotifyAuthenticator: auth,
		spotifyActivated:     false,
		UserCollection:       userCollection,
		SongCollection:       songCollection,
		SessionCollection:    sessCollection,
		Broker:               broker,
	}
}
