package db

import (
	"context"
	"fmt"

	"github.com/antonbaumann/spotify-jukebox/config"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type Model struct {
	Client   *mongo.Client
	Database *mongo.Database
}

func New(ctx context.Context) (*Model, error) {
	mongoURI := fmt.Sprintf(
		"mongodb://%v:%v",
		config.Conf.Database.DBHost,
		config.Conf.Database.DBPort,
	)
	clientOptions := options.Client().ApplyURI(mongoURI)

	client, err := mongo.Connect(ctx, clientOptions)
	if err != nil {
		return nil, err
	}

	// Check the connection
	err = client.Ping(ctx, nil)
	if err != nil {
		return nil, err
	}

	return &Model{
		Client: client,
		Database: client.Database(config.Conf.Database.DBName),
	}, nil
}
