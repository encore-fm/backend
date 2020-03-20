package server

import (
	"github.com/antonbaumann/spotify-jukebox/config"
	"github.com/antonbaumann/spotify-jukebox/db"
	"github.com/antonbaumann/spotify-jukebox/handlers"
)

type Model struct {
	Port          int
	UserHandler   *handlers.UserHandler
	AdminHandler  *handlers.AdminHandler
	ServerHandler *handlers.ServerHandler
}

func New(dbConn *db.Model) *Model {
	serverHandler := &handlers.ServerHandler{}

	adminHandler := &handlers.AdminHandler{
		UserCollection: db.NewUserCollection(dbConn.Client),
	}

	userHandler := &handlers.UserHandler{
		UserCollection: db.NewUserCollection(dbConn.Client),
	}

	server := &Model{
		Port:          config.Conf.Server.Port,
		UserHandler:   userHandler,
		ServerHandler: serverHandler,
		AdminHandler:  adminHandler,
	}

	return server
}
