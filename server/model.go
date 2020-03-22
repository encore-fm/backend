package server

import (
	"github.com/antonbaumann/spotify-jukebox/config"
	"github.com/antonbaumann/spotify-jukebox/db"
	"github.com/antonbaumann/spotify-jukebox/handlers"
	"github.com/zmb3/spotify"
)

type Model struct {
	Port    int
	Handler *handlers.Handler
}

func New(dbConn *db.Model, spotifyAuth spotify.Authenticator) *Model {
	handler := handlers.New(
		db.NewUserCollection(dbConn.Client),
		db.NewSongCollection(dbConn.Client),
		spotifyAuth,
	)

	server := &Model{
		Port:    config.Conf.Server.Port,
		Handler: handler,
	}

	return server
}
