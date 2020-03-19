package server

import (
	"github.com/antonbaumann/spotify-jukebox/config"
	"github.com/antonbaumann/spotify-jukebox/db"
	"github.com/antonbaumann/spotify-jukebox/handlers"
)

type Model struct {
	Port          int
	DBConn        *db.Model
	UserHandler   *handlers.UserHandler
	ServerHandler *handlers.ServerHandler
}

func New(dbConn *db.Model) *Model {
	serverHandler := &handlers.ServerHandler{}

	userHandler := &handlers.UserHandler{
		UserCollection: db.NewUserCollection(dbConn.Client),
	}

	server := &Model{
		Port:          config.Conf.Server.Port,
		DBConn:        dbConn,
		UserHandler:   userHandler,
		ServerHandler: serverHandler,
	}

	return server
}
