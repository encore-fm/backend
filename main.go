package main

import (
	"context"

	"github.com/antonbaumann/spotify-jukebox/config"
	"github.com/antonbaumann/spotify-jukebox/db"
	"github.com/antonbaumann/spotify-jukebox/server"
	"github.com/antonbaumann/spotify-jukebox/sse"
	log "github.com/sirupsen/logrus"
	"github.com/zmb3/spotify"
	"golang.org/x/oauth2/clientcredentials"
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

	// create spotify authenticator
	spotifyAuth := spotifyAuthSetup()

	// create spotify client
	spotifyConf := &clientcredentials.Config{
		ClientID:     config.Conf.Spotify.ClientID,
		ClientSecret: config.Conf.Spotify.ClientSecret,
		TokenURL:     spotify.TokenURL,
	}
	token, err := spotifyConf.Token(context.Background())
	if err != nil {
		log.Fatalf("[startup] couldn't get spotify authentication token: %v", err)
	}
	spotifyClient := spotify.Authenticator{}.NewClient(token)
	log.Info("[startup] successfully connected to spotify api")

	// create broker
	sseBroker := sse.NewBroker()
	sseBroker.Start()
	log.Info("[startup] successfully started SSE broker")

	// start server
	svr := server.New(dbConn, spotifyAuth, spotifyClient, sseBroker)
	svr.Start()
}
