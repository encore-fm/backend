package systest

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/antonbaumann/spotify-jukebox/config"
	"github.com/antonbaumann/spotify-jukebox/db"
	"github.com/antonbaumann/spotify-jukebox/handlers"
	"github.com/antonbaumann/spotify-jukebox/session"
	"github.com/antonbaumann/spotify-jukebox/song"
	"github.com/antonbaumann/spotify-jukebox/user"
	"github.com/antonbaumann/spotify-jukebox/util"
	"github.com/stretchr/testify/assert"
	"go.mongodb.org/mongo-driver/mongo"
	"io/ioutil"
	"net/http"
	"os"
	"testing"
)

const (
	BackendBaseUrl    = "http://127.0.0.1:8080"
	TestAdminUsername = "baumanto"
	TestAdminSecret   = "1234"
	TestSessionID     = "91cb20eeb38943efb0ee22147f3a1b70d59d4d8e3b3c54771bb89ea8cce3f0c51b8e6ef4faab22e7f1c9e3db9c8c220eb26f5ed8936a63b63727648ea9963698"
	NotRickRollSongID = "4uLU6hMCjMI75M1A2tKUQC"
)

var (
	sessionCollection *mongo.Collection
	userCollection    *mongo.Collection

	testSession = &session.Session{
		ID:       TestSessionID,
		SongList: make([]*song.Model, 0),
	}
	testAdmin = &user.Model{
		ID:                fmt.Sprintf("%v@%v", TestAdminUsername, TestSessionID),
		Username:          TestAdminUsername,
		Secret:            TestAdminSecret,
		SessionID:         TestSessionID,
		IsAdmin:           true,
		Score:             1,
		SpotifyAuthorized: false,
		AuthToken:         nil,
		AuthState:         "4321",
	}
)

func TestMain(m *testing.M) {
	// Create and open db connection
	dbConn, err := db.New() // todo maybe write create and open db myself to avoid db package dependency
	if err != nil {
		panic(err)
	}
	sessionCollection = dbConn.Client.
		Database(config.Conf.Database.DBName).
		Collection(config.Conf.Database.SessionCollectionName)
	userCollection = dbConn.Client.
		Database(config.Conf.Database.DBName).
		Collection(config.Conf.Database.UserCollectionName)

	// Add test data
	_, err = sessionCollection.InsertOne(context.Background(), testSession)
	if err != nil {
		panic(err)
	}
	_, err = userCollection.InsertOne(context.Background(), testAdmin)
	if err != nil {
		panic(err)
	}

	status := m.Run()

	// drop database
	err = dbConn.Client.Database(config.Conf.Database.DBName).Drop(context.Background())
	if err != nil {
		panic(err)
	}

	os.Exit(status)
}

// Tests adding a new user to an existing session. Expects normal behavior.
func Test_UserJoin_ExistingSession(t *testing.T) {
	username := "jonhue"
	sessionID := TestSessionID

	count, err := userCollection.CountDocuments(context.Background(), nil)
	if err != nil {
		t.Fatal(err)
	}

	endpointUrl := fmt.Sprintf("%v//users/%v/join/%v", BackendBaseUrl, username, sessionID)

	resp, err := http.Post(endpointUrl, "", nil)
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()

	assert.Equal(t, http.StatusOK, resp.StatusCode)

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		t.Fatal(err)
	}

	response := &struct {
		UserInfo *user.Model `json:"user_info"`
		AuthUrl  string      `json:"auth_url"`
	}{}

	err = json.Unmarshal(body, response)
	if err != nil {
		t.Fatal(err)
	}

	var foundUser *user.Model
	err = userCollection.FindOne(context.Background(), response.UserInfo).Decode(foundUser)
	if err != nil {
		t.Fatal(err)
	}

	newCount, err := userCollection.CountDocuments(context.Background(), nil)
	if err != nil {
		t.Fatal(err)
	}

	assert.Equal(t, response.UserInfo, foundUser) // make sure user is written into db
	assert.Equal(t, count+1, newCount)            // make sure the db only added one document to usercollection
	assert.Equal(t, username, response.UserInfo.Username)
	assert.Equal(t, TestSessionID, response.UserInfo.SessionID)
	assert.Equal(t, 1, response.UserInfo.Score)
	assert.Equal(t, false, response.UserInfo.IsAdmin)
}

func Test_UserJoin_NonExistingSession(t *testing.T) {
	username := "eti"
	sessionID, err := util.GenerateSecret()
	if err != nil {
		t.Fatal(err)
	}

	count, err := userCollection.CountDocuments(context.Background(), nil)
	if err != nil {
		t.Fatal(err)
	}

	endpointUrl := fmt.Sprintf("%v/users/%v/join/%v", BackendBaseUrl, username, sessionID)

	resp, err := http.Post(endpointUrl, "", nil)
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()

	assert.Equal(t, http.StatusNotFound, resp.StatusCode) // response should not be okay

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		t.Fatal(err)
	}

	response := handlers.FrontendError{}

	err = json.Unmarshal(body, response)
	if err != nil {
		t.Fatal(err)
	}

	newCount, err := userCollection.CountDocuments(context.Background(), nil)
	if err != nil {
		t.Fatal(err)
	}

	assert.Equal(t, handlers.SessionNotFoundError, response) // make sure the correct frontenderror is returned
	assert.Equal(t, count, newCount)                         // make sure no new documents were added to usercollection
}

func Test_UserJoin_ExistingUser(t *testing.T) {
	username := TestAdminUsername
	sessionID := TestSessionID

	count, err := userCollection.CountDocuments(context.Background(), nil)
	if err != nil {
		t.Fatal(err)
	}

	endpointUrl := fmt.Sprintf("%v/users/%v/join/%v", BackendBaseUrl, username, sessionID)

	resp, err := http.Post(endpointUrl, "", nil)
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()

	assert.Equal(t, http.StatusConflict, resp.StatusCode) // 409 expected when username exists

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		t.Fatal(err)
	}

	response := handlers.FrontendError{}

	err = json.Unmarshal(body, response)
	if err != nil {
		t.Fatal(err)
	}

	newCount, err := userCollection.CountDocuments(context.Background(), nil)
	if err != nil {
		t.Fatal(err)
	}

	assert.Equal(t, handlers.UserConflictError, response)
	assert.Equal(t, count, newCount)
}

func Test_UserList(t *testing.T) {
	username := TestAdminUsername

	count, err := userCollection.CountDocuments(context.Background(), nil)
	if err != nil {
		t.Fatal(err)
	}

	endpointUrl := fmt.Sprintf("%v/users/%v/list", BackendBaseUrl, username)

	resp, err := http.Get(endpointUrl)
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		t.Fatal(err)
	}

	var response []*user.ListElement

	err = json.Unmarshal(body, response)
	if err != nil {
		t.Fatal(err)
	}

	assert.Equal(t, count, len(response))
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func Test_UserSuggestSong_GoodID(t *testing.T) {
	username := TestAdminUsername
	songID := NotRickRollSongID

	endpointUrl := fmt.Sprintf("%v/users/%v/suggest/%v", BackendBaseUrl, username, songID)

	resp, err := http.Post(endpointUrl, "", nil)
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()

	assert.Equal(t, http.StatusOK, resp.StatusCode)

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		t.Fatal(err)
	}

	response := song.Model{}

	err = json.Unmarshal(body, response)
	if err != nil {
		t.Fatal(err)
	}

	assert.Equal(t, songID, response.ID)
	assert.Equal(t, 1, response.Score) // song score should be 1 after being suggested
	assert.Equal(t, 1, len(response.Upvoters))
	assert.True(t, util.Contains(response.Upvoters, username))
}

func Test_UserSuggestSong_BadID(t *testing.T) {
	username := TestAdminUsername
	songID := ""

	endpointUrl := fmt.Sprintf("%v/users/%v/suggest/%v", BackendBaseUrl, username, songID)

	resp, err := http.Post(endpointUrl, "", nil)
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()

	assert.NotEqual(t, http.StatusOK, resp.StatusCode)
}

func Test_UserSuggestSong_BadUser(t *testing.T) {
	username := ""
	songID := NotRickRollSongID

	endpointUrl := fmt.Sprintf("%v/users/%v/suggest/%v", BackendBaseUrl, username, songID)

	resp, err := http.Post(endpointUrl, "", nil)
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()

	assert.NotEqual(t, http.StatusOK, resp.StatusCode)
}
