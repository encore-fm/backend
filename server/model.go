package server

import (
	"github.com/encore-fm/backend/config"
	"github.com/encore-fm/backend/db"
	"github.com/encore-fm/backend/events"
	"github.com/encore-fm/backend/handlers"
	"github.com/encore-fm/backend/spotifycl"
	"github.com/zmb3/spotify"
)

type Model struct {
	Port              int
	UserCollection    db.UserCollection
	SessionCollection db.SessionCollection
	SongCollection    db.SongCollection
	SSEHandler        handlers.SSEHandler
	AdminHandler      handlers.AdminHandler
	UserHandler       handlers.UserHandler
	ServerHandler     handlers.ServerHandler
	SpotifyHandler    handlers.SpotifyHandler
	PlayerHandler     handlers.PlayerHandler
	DebugHandler      handlers.DebugHandler
	EventBus          events.EventBus
}

func New(
	eventBus events.EventBus,
	userHandle db.UserCollection,
	sessHandle db.SessionCollection,
	songHandle db.SongCollection,
	playerHandle db.PlayerCollection,
	spotifyAuth spotify.Authenticator,
	spotifyClient *spotifycl.SpotifyClient,
) *Model {

	handler := handlers.New(
		eventBus,
		userHandle,
		sessHandle,
		songHandle,
		playerHandle,
		spotifyAuth,
		spotifyClient,
	)

	server := &Model{
		Port:              config.Conf.Server.Port,
		UserCollection:    userHandle,
		SessionCollection: sessHandle,
		SongCollection:    songHandle,
		SSEHandler:        handlers.SSEHandler(handler),
		AdminHandler:      handlers.AdminHandler(handler),
		UserHandler:       handlers.UserHandler(handler),
		ServerHandler:     handlers.ServerHandler(handler),
		SpotifyHandler:    handlers.SpotifyHandler(handler),
		PlayerHandler:     handlers.PlayerHandler(handler),
		DebugHandler:      handlers.DebugHandler(handler),
		EventBus:          eventBus,
	}

	return server
}
