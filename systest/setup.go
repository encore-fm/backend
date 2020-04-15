package systest

import (
	"context"
	"fmt"
	"net/http"

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

func resetPlayerController(sessionID string) (*http.Response, error) {
	endpointUrl := fmt.Sprintf("%v/debug/reset_player_controller/%v", BackendBaseUrl, sessionID)

	client := &http.Client{}
	req, err := http.NewRequest("POST", endpointUrl, nil)
	if err != nil {
		return nil, err
	}

	return client.Do(req)
}
