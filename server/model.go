package server

import (
	"github.com/antonbaumann/spotify-jukebox/config"
	"github.com/antonbaumann/spotify-jukebox/db"
	"github.com/antonbaumann/spotify-jukebox/handlers"
	"github.com/zmb3/spotify"
)

type Model struct {
	Port           int
	UserCollection db.UserCollection
	SongCollection db.SongCollection
	AdminHandler   handlers.AdminHandler
	UserHandler    handlers.UserHandler
	ServerHandler  handlers.ServerHandler
	SpotifyHandler handlers.SpotifyHandler
}

func New(dbConn *db.Model, spotifyAuth spotify.Authenticator) *Model {
	userHandle := db.NewUserCollection(dbConn.Client)
	songHandle := db.NewSongCollection(dbConn.Client)

	handler := handlers.New(
		userHandle,
		songHandle,
		spotifyAuth,
	)

	server := &Model{
		Port:           config.Conf.Server.Port,
		UserCollection: userHandle,
		SongCollection: songHandle,
		AdminHandler:   handlers.AdminHandler(handler),
		UserHandler:    handlers.UserHandler(handler),
		ServerHandler:  handlers.ServerHandler(handler),
		SpotifyHandler: handlers.SpotifyHandler(handler),
	}

	return server
}
