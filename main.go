package main

import (
	"context"

	"github.com/antonbaumann/spotify-jukebox/config"
	"github.com/antonbaumann/spotify-jukebox/db"
	"github.com/antonbaumann/spotify-jukebox/server"
	log "github.com/sirupsen/logrus"
	"github.com/zmb3/spotify"
)

func main() {
	// connect to database
	dbConn, err := db.New(context.TODO())
	if err != nil {
		panic(err)
	}
	log.Infof(
		"successfully connected to database at %v:%v",
		config.Conf.Database.DBHost,
		config.Conf.Database.DBPort,
	)

	// create spotify client
	spotifyAuth := spotify.NewAuthenticator(
		config.Conf.Spotify.RedirectUrl,
		spotify.ScopeUserReadPrivate,
	)
	spotifyAuth.SetAuthInfo(
		config.Conf.Spotify.ClientID,
		config.Conf.Spotify.ClientSecret,
	)
	url := spotifyAuth.AuthURL(config.Conf.Spotify.State)
	log.Infof("Go to %v", url)

	// start server
	svr := server.New(dbConn, spotifyAuth)
	svr.Start()
}
