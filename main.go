package main

import (
	"github.com/antonbaumann/spotify-jukebox/config"
	"github.com/antonbaumann/spotify-jukebox/db"
	"github.com/antonbaumann/spotify-jukebox/server"
	"github.com/antonbaumann/spotify-jukebox/spotifycl"
	"github.com/antonbaumann/spotify-jukebox/sse"
	log "github.com/sirupsen/logrus"
	"github.com/zmb3/spotify"
)

func spotifyAuthSetup() spotify.Authenticator {
	spotifyAuth := spotify.NewAuthenticator(
		config.Conf.Spotify.RedirectUrl,
		spotify.ScopeUserReadPrivate,
	)
	spotifyAuth.SetAuthInfo(
		config.Conf.Spotify.ClientID,
		config.Conf.Spotify.ClientSecret,
	)
	return spotifyAuth
}

func main() {
	config.Setup()
	// connect to database
	dbConn, err := db.New()
	if err != nil {
		panic(err)
	}
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

	// create broker
	sseBroker := sse.NewBroker()
	sseBroker.Start()
	log.Info("[startup] successfully started SSE broker")

	// start server
	svr := server.New(dbConn, spotifyAuth, spotifyClient, sseBroker)
	svr.Start()
}
