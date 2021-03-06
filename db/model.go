package db

import (
	"context"
	"fmt"
	"time"

	"github.com/encore-fm/backend/config"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type Model struct {
	Client *mongo.Client
}

func New() (*Model, error) {
	mongoURI := fmt.Sprintf(
		"mongodb+srv://%v:%v@%v/test?retryWrites=true&w=majority",
		config.Conf.Database.DBUser,
		config.Conf.Database.DBPassword,
		config.Conf.Database.DBHost,
	)

	if config.Conf.Server.Debug {
		mongoURI = fmt.Sprintf(
			"mongodb://%v:%v@%v:%v/?connect=direct",
			config.Conf.Database.DBUser,
			config.Conf.Database.DBPassword,
			config.Conf.Database.DBHost,
			config.Conf.Database.DBPort,
		)
	}

	clientOptions := options.Client().ApplyURI(mongoURI)

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
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
	}, nil
}
