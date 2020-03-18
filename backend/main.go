package main

import (
	"github.com/antonbaumann/spotify-jukebox/config"
	"github.com/antonbaumann/spotify-jukebox/db"
	"github.com/antonbaumann/spotify-jukebox/server"
	log "github.com/sirupsen/logrus"
)

func main() {
	// connect to database
	client, err := db.Connect()
	if err != nil {
		panic(err)
	}
	log.Infof(
		"successfully connected to database at %v:%v",
		config.Conf.Database.DBHost,
		config.Conf.Database.DBPort,
	)

	// start server
	svr := server.New(client)
	svr.Start()
}
