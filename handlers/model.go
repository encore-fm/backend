package handlers

import (
	"github.com/antonbaumann/spotify-jukebox/db"
	"github.com/antonbaumann/spotify-jukebox/events"
	"github.com/antonbaumann/spotify-jukebox/spotifycl"
	"github.com/zmb3/spotify"
)

type handler struct {
	eventBus             events.EventBus
	spotifyAuthenticator spotify.Authenticator
	Spotify              *spotifycl.SpotifyClient
	UserCollection       db.UserCollection
	SessionCollection    db.SessionCollection
}

func New(
	eventBus events.EventBus,
	userCollection db.UserCollection,
	sessCollection db.SessionCollection,
	auth spotify.Authenticator,
	client *spotifycl.SpotifyClient,
) *handler {
	return &handler{
		eventBus:             eventBus,
		spotifyAuthenticator: auth,
		Spotify:              client,
		UserCollection:       userCollection,
		SessionCollection:    sessCollection,
	}
}
