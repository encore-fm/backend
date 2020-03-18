package db

import (
	"context"
	"fmt"

	"github.com/antonbaumann/spotify-jukebox/config"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func Connect() (*mongo.Client, error) {
	mongoURI := fmt.Sprintf(
		"mongodb://%v:%v",
		config.Conf.Database.DBHost,
		config.Conf.Database.DBPort,
	)
	clientOptions := options.Client().ApplyURI(mongoURI)
	return mongo.Connect(context.Background(), clientOptions)
}
