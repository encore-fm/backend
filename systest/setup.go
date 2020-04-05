package systest

import (
	"context"

	"github.com/antonbaumann/spotify-jukebox/config"
)

func setupDB() {
	// setup collections
	sessionCollection = client.
		Database(config.Conf.Database.DBName).
		Collection(config.Conf.Database.SessionCollectionName)
	userCollection = client.
		Database(config.Conf.Database.DBName).
		Collection(config.Conf.Database.UserCollectionName)

	// Add test data
	_, err := sessionCollection.InsertOne(context.Background(), testSession)
	if err != nil {
		panic(err)
	}
	_, err = userCollection.InsertOne(context.Background(), testAdmin)
	if err != nil {
		panic(err)
	}
	_, err = userCollection.InsertOne(context.Background(), testUser)
	if err != nil {
		panic(err)
	}
}

func dropDB() {
	err := client.Database(config.Conf.Database.DBName).Drop(context.Background())
	if err != nil {
		panic(err)
	}
}