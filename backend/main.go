package main

import (
	"github.com/antonbaumann/spotify-jukebox/config"
	"github.com/antonbaumann/spotify-jukebox/server"
)

func main() {
	svr := server.New(config.Conf.Server.Port)
	svr.Start()
}