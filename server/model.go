package server

import (
	"github.com/antonbaumann/spotify-jukebox/config"
	"github.com/antonbaumann/spotify-jukebox/db"
	"github.com/antonbaumann/spotify-jukebox/handlers"
	"github.com/zmb3/spotify"
)

type Model struct {
	Port           int
	UserHandler    *handlers.UserHandler
	AdminHandler   *handlers.AdminHandler
	ServerHandler  *handlers.ServerHandler
	SpotifyHandler *handlers.SpotifyHandler
}

func New(dbConn *db.Model, spotifyAuth spotify.Authenticator) *Model {
	serverHandler := &handlers.ServerHandler{}

	spotifyHandler := &handlers.SpotifyHandler{
		Authenticator: spotifyAuth,
	}

	adminHandler := &handlers.AdminHandler{
		UserCollection: db.NewUserCollection(dbConn.Client),
	}

	userHandler := &handlers.UserHandler{
		UserCollection: db.NewUserCollection(dbConn.Client),
	}

	server := &Model{
		Port:           config.Conf.Server.Port,
		UserHandler:    userHandler,
		ServerHandler:  serverHandler,
		AdminHandler:   adminHandler,
		SpotifyHandler: spotifyHandler,
	}

	return server
}
