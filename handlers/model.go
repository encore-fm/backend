package handlers

import (
	"github.com/encore-fm/backend/db"
	"github.com/encore-fm/backend/events"
	"github.com/encore-fm/backend/spotifycl"
	"github.com/zmb3/spotify"
)

type handler struct {
	eventBus             events.EventBus
	spotifyAuthenticator spotify.Authenticator
	Spotify              *spotifycl.SpotifyClient
	UserCollection       db.UserCollection
	SessionCollection    db.SessionCollection
	SongCollection       db.SongCollection
	PlayerCollection     db.PlayerCollection
}

func New(
	eventBus events.EventBus,
	userCollection db.UserCollection,
	sessCollection db.SessionCollection,
	songCollection db.SongCollection,
	playerCollection db.PlayerCollection,
	auth spotify.Authenticator,
	client *spotifycl.SpotifyClient,
) *handler {
	return &handler{
		eventBus:             eventBus,
		spotifyAuthenticator: auth,
		Spotify:              client,
		UserCollection:       userCollection,
		SessionCollection:    sessCollection,
		SongCollection:       songCollection,
		PlayerCollection:     playerCollection,
	}
}
