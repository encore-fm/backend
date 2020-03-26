package handlers

import (
	"github.com/antonbaumann/spotify-jukebox/db"
	"github.com/zmb3/spotify"
)

type handler struct {
	spotifyAuthenticator spotify.Authenticator
	Spotify              spotify.Client
	spotifyActivated     bool
	UserCollection       db.UserCollection
	SongCollection       db.SongCollection
}

func New(
	userCollection db.UserCollection,
	songCollection db.SongCollection,
	auth spotify.Authenticator,
) *handler {
	return &handler{
		spotifyAuthenticator: auth,
		spotifyActivated:     false,
		UserCollection:       userCollection,
		SongCollection:       songCollection,
	}
}
