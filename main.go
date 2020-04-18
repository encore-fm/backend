package main

import (
	"github.com/antonbaumann/spotify-jukebox/config"
	"github.com/antonbaumann/spotify-jukebox/db"
	"github.com/antonbaumann/spotify-jukebox/events"
	"github.com/antonbaumann/spotify-jukebox/playerctrl"
	"github.com/antonbaumann/spotify-jukebox/server"
	"github.com/antonbaumann/spotify-jukebox/spotifycl"
	log "github.com/sirupsen/logrus"
	"github.com/zmb3/spotify"
)

func spotifyAuthSetup() spotify.Authenticator {
	spotifyAuth := spotify.NewAuthenticator(
		config.Conf.Spotify.RedirectUrl,
		spotify.ScopeStreaming,
		spotify.ScopeUserReadEmail,
		spotify.ScopeUserModifyPlaybackState,
		spotify.ScopeUserReadPrivate,
		spotify.ScopeUserReadPlaybackState,
	)
	spotifyAuth.SetAuthInfo(
		config.Conf.Spotify.ClientID,
		config.Conf.Spotify.ClientSecret,
	)
	return spotifyAuth
}

func main() {
	config.Setup()

	// init event bus
	eventBus := events.NewEventBus()
	eventBus.Start()

	// connect to database
	dbConn, err := db.New()
	if err != nil {
		panic(err)
	}
	userDB := db.NewUserCollection(dbConn.Client)
	sessDB := db.NewSessionCollection(dbConn.Client)
	songDB := db.NewSongCollection(dbConn.Client)
	playerDB := db.NewPlayerCollection(dbConn.Client)
	log.Infof(
		"[startup] successfully connected to database at %v:%v",
		config.Conf.Database.DBHost,
		config.Conf.Database.DBPort,
	)

	// create spotify client
	spotifyClient, err := spotifycl.New(config.Conf.Spotify.ClientID, config.Conf.Spotify.ClientSecret)
	if err != nil {
		log.Fatalf("[startup] creating spotify client: %v", err)
	}
	spotifyClient.Start()
	log.Info("[startup] successfully connected to spotify api")

	// create spotify authenticator
	spotifyAuth := spotifyAuthSetup()

	// create controller
	playerCtrl := playerctrl.NewController(
		eventBus,
		sessDB,
		songDB,
		userDB,
		playerDB,
		spotifyAuth,
	)
	if err := playerCtrl.Start(); err != nil {
		log.Fatalf("[startup] starting player controller: %v", err)
	}
	log.Info("[startup] successfully started player controller")

	// start server
	svr := server.New(eventBus, userDB, sessDB, songDB, playerDB, spotifyAuth, spotifyClient)
	svr.Start()
}
