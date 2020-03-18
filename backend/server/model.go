package server

import (
	"github.com/antonbaumann/spotify-jukebox/config"
	"go.mongodb.org/mongo-driver/mongo"
)

type Model struct {
	Port     int
	DBClient *mongo.Client
}

func New(client *mongo.Client) *Model {
	server := &Model{
		Port:     config.Conf.Server.Port,
		DBClient: client,
	}
	return server
}
