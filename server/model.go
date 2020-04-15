package server

import (
	"github.com/antonbaumann/spotify-jukebox/config"
	"github.com/antonbaumann/spotify-jukebox/db"
	"github.com/antonbaumann/spotify-jukebox/events"
	"github.com/antonbaumann/spotify-jukebox/handlers"
	"github.com/antonbaumann/spotify-jukebox/spotifycl"
	"github.com/antonbaumann/spotify-jukebox/sse"
	"github.com/zmb3/spotify"
)

type Model struct {
	Port              int
	UserCollection    db.UserCollection
	SessionCollection db.SessionCollection
	SongCollection    db.SongCollection
	SSEHandler        sse.Handler
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

	sseHandler := sse.New(eventBus)

	server := &Model{
		Port:              config.Conf.Server.Port,
		UserCollection:    userHandle,
		SessionCollection: sessHandle,
		SongCollection:    songHandle,
		SSEHandler:        sseHandler,
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
