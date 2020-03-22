package handlers

import (
	"github.com/antonbaumann/spotify-jukebox/db"
	"github.com/zmb3/spotify"
)

type Handler struct {
	spotifyAuthenticator spotify.Authenticator
	Spotify              spotify.Client
	UserCollection       *db.UserCollection
}

func New(userCollection *db.UserCollection, auth spotify.Authenticator) *Handler {
	return &Handler{
		spotifyAuthenticator:auth,
		UserCollection: userCollection,
	}
}
