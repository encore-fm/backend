package server

import (
	"github.com/antonbaumann/spotify-jukebox/config"
	"github.com/antonbaumann/spotify-jukebox/db"
	"github.com/antonbaumann/spotify-jukebox/handlers"
	"github.com/antonbaumann/spotify-jukebox/player"
	"github.com/antonbaumann/spotify-jukebox/spotifycl"
	"github.com/antonbaumann/spotify-jukebox/sse"
	"github.com/zmb3/spotify"
)

type Model struct {
	Port              int
	UserCollection    db.UserCollection
	SessionCollection db.SessionCollection
	AdminHandler      handlers.AdminHandler
	UserHandler       handlers.UserHandler
	ServerHandler     handlers.ServerHandler
	SpotifyHandler    handlers.SpotifyHandler
	Broker            *sse.Broker
}

func New(
	dbConn *db.Model,
	spotifyAuth spotify.Authenticator,
	spotifyClient *spotifycl.SpotifyClient,
	broker *sse.Broker,
	playerCtrl *player.Controller,
) *Model {
	userHandle := db.NewUserCollection(dbConn.Client)
	sessHandle := db.NewSessionCollection(dbConn.Client)

	handler := handlers.New(
		userHandle,
		sessHandle,
		spotifyAuth,
		spotifyClient,
		broker,
		playerCtrl,
	)

	server := &Model{
		Port:              config.Conf.Server.Port,
		UserCollection:    userHandle,
		SessionCollection: sessHandle,
		AdminHandler:      handlers.AdminHandler(handler),
		UserHandler:       handlers.UserHandler(handler),
		ServerHandler:     handlers.ServerHandler(handler),
		SpotifyHandler:    handlers.SpotifyHandler(handler),
		Broker:            broker,
	}

	return server
}
