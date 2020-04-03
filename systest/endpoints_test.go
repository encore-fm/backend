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
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"io/ioutil"
	"net/http"
	"os"
	"testing"
)

const (
	TestSessionID = "91cb20eeb38943efb0ee22147f3a1b70d59d4d8e3b3c54771bb89ea8cce3f0c51b8e6ef4faab22e7f1c9e3db9c8c220eb26f5ed8936a63b63727648ea9963698"

	TestAdminUsername = "baumanto"
	TestAdminSecret   = "1234"

	TestUserName   = "omar"
	TestUserSecret = "1234"

	NotRickRollSongID = "4uLU6hMCjMI75M1A2tKUQC"
	SkiFoanID         = "3gnB6G7MCqB0xjYiAdiaSY"
	AntonAusTirolID   = "2YuKyP77pidQlkxm8PuyJj"
	CordulaSongID     = "483ykNWhSQXYprueXUMNeo"
)

var (
	client *mongo.Client

	sessionCollection *mongo.Collection
	userCollection    *mongo.Collection

	testSession = &session.Session{
		ID: TestSessionID,
		SongList: []*song.Model{{
			ID:          SkiFoanID,
			SuggestedBy: TestAdminUsername,
			Score:       1,
			Upvoters:    []string{TestAdminUsername},
			Downvoters:  make([]string, 0),
		}, {
			ID:          AntonAusTirolID,
			SuggestedBy: TestUserName,
			Score:       1,
			Upvoters:    []string{TestAdminUsername, TestUserName},
			Downvoters:  []string{},
		}, {
			ID:          CordulaSongID,
			SuggestedBy: TestAdminUsername,
			Score:       1,
			Upvoters:    []string{TestAdminUsername},
			Downvoters:  []string{TestUserName},
		}},
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
		AuthState:         fmt.Sprintf("%v:%v", TestAdminUsername, TestAdminSecret),
	}
	testUser = &user.Model{
		ID:                fmt.Sprintf("%v@%v", TestUserName, TestSessionID),
		Username:          TestUserName,
		Secret:            TestUserSecret,
		SessionID:         TestSessionID,
		IsAdmin:           false,
		Score:             1,
		SpotifyAuthorized: false,
		AuthToken:         nil,
		AuthState:         fmt.Sprintf("%v:%v", TestUserName, TestUserSecret),
	}
)

func dropDB() {
	err := client.Database(config.Conf.Database.DBName).Drop(context.Background())
	if err != nil {
		panic(err)
	}
}

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

func TestMain(m *testing.M) {
	// Create and open db connection
	config.Setup()
	dbConn, err := db.New() // todo maybe write create and open db myself to avoid db package dependency
	if err != nil {
		panic(err)
	}
	client = dbConn.Client

	dropDB()
	setupDB()
	status := m.Run()

	os.Exit(status)
}

// Tests adding a new user to an existing session. Expects normal behavior.
func Test_UserJoin_ExistingSession(t *testing.T) {
	username := "jonhue"
	sessionID := TestSessionID

	// get db collection count before insertion
	count, err := userCollection.CountDocuments(context.Background(), bson.D{})
	assert.NoError(t, err)

	// post request, expect http status OK
	resp, err := UserJoin(username, sessionID)
	assert.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	// deserialize response body and assert expected results
	body, err := ioutil.ReadAll(resp.Body)
	assert.NoError(t, err)

	response := &struct {
		UserInfo *user.Model `json:"user_info"`
		AuthUrl  string      `json:"auth_url"`
	}{}

	err = json.Unmarshal(body, response)
	assert.NoError(t, err)
	assert.Equal(t, username, response.UserInfo.Username)       // make sure username matches
	assert.Equal(t, TestSessionID, response.UserInfo.SessionID) // make sure session id matches
	assert.Equal(t, 1, response.UserInfo.Score)                 // make sure score is initialized with 1
	assert.Equal(t, false, response.UserInfo.IsAdmin)           // make sure user is not admin

	// make sure db is written into db
	foundUser := &user.Model{}
	err = userCollection.FindOne(context.Background(), response.UserInfo).Decode(foundUser)
	assert.NoError(t, err)
	assert.Equal(t, response.UserInfo, foundUser)

	// get new count
	newCount, err := userCollection.CountDocuments(context.Background(), bson.D{})
	assert.NoError(t, err)
	assert.Equal(t, count+1, newCount) // make sure the db only added one document to usercollection
}

func Test_UserJoin_NonExistingSession(t *testing.T) {
	username := "eti"
	sessionID, err := util.GenerateSecret()
	assert.NoError(t, err)

	// get db collection count before insertion
	count, err := userCollection.CountDocuments(context.Background(), bson.D{})
	assert.NoError(t, err)

	// post request, expect http status not OK
	resp, err := UserJoin(username, sessionID)
	assert.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, http.StatusNotFound, resp.StatusCode) // expect 404 when session is not found

	// deserialize response body and assert expected results
	body, err := ioutil.ReadAll(resp.Body)
	assert.NoError(t, err)

	response := &handlers.FrontendError{}

	err = json.Unmarshal(body, response)
	assert.NoError(t, err)
	assert.Equal(t, handlers.SessionNotFoundError, *response) // make sure the correct frontenderror is returned

	// get new count
	newCount, err := userCollection.CountDocuments(context.Background(), bson.D{})
	assert.NoError(t, err)
	assert.Equal(t, count, newCount) // make sure no new documents were added to usercollection
}

func Test_UserJoin_ExistingUser(t *testing.T) {
	username := TestAdminUsername
	sessionID := TestSessionID

	// get db collection count before insertion
	count, err := userCollection.CountDocuments(context.Background(), bson.D{})
	assert.NoError(t, err)

	// post request, expect http status not OK
	resp, err := UserJoin(username, sessionID)
	assert.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, http.StatusConflict, resp.StatusCode) // 409 expected when username exists

	// deserialize response body and assert expected results
	body, err := ioutil.ReadAll(resp.Body)
	assert.NoError(t, err)

	response := &handlers.FrontendError{}

	err = json.Unmarshal(body, response)
	assert.NoError(t, err)
	assert.Equal(t, handlers.UserConflictError, *response)

	// get new count
	newCount, err := userCollection.CountDocuments(context.Background(), bson.D{})
	assert.NoError(t, err)
	assert.Equal(t, count, newCount) // make sure no new documents were added to usercollection
}

func Test_UserList(t *testing.T) {
	username := TestAdminUsername
	secret := TestAdminSecret
	sessionID := TestSessionID

	count, err := userCollection.CountDocuments(context.Background(), bson.D{})
	assert.NoError(t, err)

	resp, err := UserList(username, secret, sessionID)
	assert.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	body, err := ioutil.ReadAll(resp.Body)
	assert.NoError(t, err)

	response := make([]*user.ListElement, 0)

	err = json.Unmarshal(body, &response)
	assert.NoError(t, err)
	assert.Equal(t, int(count), len(response))
}

func Test_UserSuggestSong_GoodID(t *testing.T) {
	username := TestAdminUsername
	secret := TestAdminSecret
	sessionID := TestSessionID
	songID := NotRickRollSongID

	// get song list count before request
	var foundSession *session.Session
	err := sessionCollection.FindOne(context.Background(), bson.D{{"_id", sessionID}}).Decode(&foundSession)
	assert.NoError(t, err)
	count := len(foundSession.SongList)

	resp, err := UserSuggestSong(username, secret, sessionID, songID)
	assert.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusOK, resp.StatusCode)

	body, err := ioutil.ReadAll(resp.Body)
	assert.NoError(t, err)

	response := &song.Model{}

	err = json.Unmarshal(body, response)
	assert.NoError(t, err)

	assert.Equal(t, songID, response.ID)
	assert.Equal(t, 1, response.Score) // song score should be 1 after being suggested
	assert.Equal(t, 1, len(response.Upvoters))
	// make sure user is in upvoters
	index := util.Find(
		len(response.Upvoters),
		func(i int) bool {
			return response.Upvoters[i] == username
		},
	)
	assert.NotEqual(t, -1, index)

	err = sessionCollection.FindOne(context.Background(), bson.D{{"_id", sessionID}}).Decode(&foundSession)
	assert.NoError(t, err)
	newCount := len(foundSession.SongList)
	assert.Equal(t, count+1, newCount)
}

func Test_UserSuggestSong_BadID(t *testing.T) {
	username := TestAdminUsername
	secret := TestAdminSecret
	sessionID := TestSessionID
	songID := ""

	// get song list count before request
	var foundSession *session.Session
	err := sessionCollection.FindOne(context.Background(), bson.D{{"_id", sessionID}}).Decode(&foundSession)
	assert.NoError(t, err)
	count := len(foundSession.SongList)

	resp, err := UserSuggestSong(username, secret, sessionID, songID)
	assert.NoError(t, err)
	defer resp.Body.Close()

	// make sure response code is not OK
	assert.NotEqual(t, http.StatusOK, resp.StatusCode)

	// make sure no songs were added to db
	err = sessionCollection.FindOne(context.Background(), bson.D{{"_id", sessionID}}).Decode(&foundSession)
	assert.NoError(t, err)
	newCount := len(foundSession.SongList)
	assert.Equal(t, count, newCount)
}

// this function implicitly tests the user auth function associated with the user handlers
func Test_UserSuggestSong_BadUser(t *testing.T) {
	username := ""
	secret := ""
	sessionID := TestSessionID
	songID := NotRickRollSongID

	resp, err := UserSuggestSong(username, secret, sessionID, songID)
	assert.NoError(t, err)
	defer resp.Body.Close()

	assert.NotEqual(t, http.StatusOK, resp.StatusCode)
}

func Test_UserlistSongs(t *testing.T) {
	username := TestAdminUsername
	secret := TestAdminSecret
	sessionID := TestSessionID

	// get song list count before request
	var foundSession *session.Session
	err := sessionCollection.FindOne(
		context.Background(),
		bson.D{
			{"_id", sessionID},
		},
	).Decode(&foundSession)
	assert.NoError(t, err)
	count := len(foundSession.SongList)

	resp, err := UserlistSongs(username, secret, sessionID)
	assert.NoError(t, err)
	defer resp.Body.Close()

	// make sure response code is OK
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	body, err := ioutil.ReadAll(resp.Body)
	assert.NoError(t, err)

	response := make([]*song.Model, 0)

	err = json.Unmarshal(body, &response)
	assert.NoError(t, err)

	// make sure count matches response
	assert.Equal(t, count, len(response))
}

// Case 1: user neither upvoter nor downvoter: add to upvoters, increment score by 1
// This test uses the song Skifoan, which was suggested by the admin and neither upvoted nor downvoted by the user
// The test sends an upvote request by the user
func Test_UserUpvote_Case1(t *testing.T) {
	dropDB()
	setupDB()

	username := TestUserName
	secret := TestUserSecret
	sessionID := TestSessionID
	songID := SkiFoanID

	// get upvoters and downvoters
	filter := bson.D{
		{"_id", sessionID},
		{"song_list.id", songID},
	}
	projection := bson.D{
		{"_id", 0},
		{"song_list.$", 1},
	}
	var sess *session.Session
	err := sessionCollection.FindOne(
		context.Background(),
		filter,
		options.FindOne().SetProjection(projection),
	).Decode(&sess)
	assert.NoError(t, err)
	songList := sess.SongList
	assert.NotEmpty(t, songList)
	upvoters := songList[0].Upvoters
	downvoters := songList[0].Downvoters

	// get suggesting user
	var suggestingUser *user.Model
	err = userCollection.FindOne(
		context.Background(),
		bson.D{
			{"_id", fmt.Sprintf("%v@%v", songList[0].SuggestedBy, sessionID)},
		},
	).Decode(&suggestingUser)
	assert.NoError(t, err)
	oldScore := suggestingUser.Score

	// make sure user is neither in upvoters nor in downvoters
	assert.Equal(t, -1, util.Find(len(upvoters),
		func(i int) bool {
			return upvoters[i] == username
		}))
	assert.Equal(t, -1, util.Find(len(downvoters),
		func(i int) bool {
			return downvoters[i] == username
		}))

	// request upvote
	resp, err := UserUpvote(username, secret, sessionID, songID)
	assert.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusOK, resp.StatusCode)

	// make sure admin score was incremented
	err = userCollection.FindOne(
		context.Background(),
		bson.D{
			{"_id", fmt.Sprintf("%v@%v", songList[0].SuggestedBy, sessionID)},
		},
	).Decode(&suggestingUser)
	assert.NoError(t, err)
	newScore := suggestingUser.Score

	assert.Equal(t, oldScore+1, newScore)

	// make sure user is in upvoters but not in downvoters
	err = sessionCollection.FindOne(context.Background(), filter, options.FindOne().SetProjection(projection)).
		Decode(&sess)
	assert.NoError(t, err)
	songList = sess.SongList
	assert.NotEmpty(t, songList)
	upvoters = songList[0].Upvoters
	downvoters = songList[0].Downvoters

	assert.NotEqual(t, -1, util.Find(len(upvoters),
		func(i int) bool {
			return upvoters[i] == username
		}))
	assert.Equal(t, -1, util.Find(len(downvoters),
		func(i int) bool {
			return downvoters[i] == username
		}))
}

// Case 2: admin in upvoters and not in downvoters, remove from upvoters, decrement score by 1
// This test uses the song Anton aus Tirol, which was suggested by the user and upvoted by the admin.
// The test sends an upvote request by the admin
func Test_UserUpvote_Case2(t *testing.T) {
	dropDB()
	setupDB()

	username := TestAdminUsername
	secret := TestAdminSecret
	sessionID := TestSessionID
	songID := AntonAusTirolID

	// get upvoters and downvoters
	filter := bson.D{
		{"_id", sessionID},
		{"song_list.id", songID},
	}
	projection := bson.D{
		{"_id", 0},
		{"song_list.$", 1},
	}
	var sess *session.Session
	err := sessionCollection.FindOne(context.Background(),
		filter,
		options.FindOne().SetProjection(projection)).
		Decode(&sess)
	assert.NoError(t, err)
	songList := sess.SongList
	assert.NotEmpty(t, songList)
	upvoters := songList[0].Upvoters
	downvoters := songList[0].Downvoters

	// get suggesting user
	var suggestingUser *user.Model
	err = userCollection.FindOne(context.Background(),
		bson.D{
			{"_id", fmt.Sprintf("%v@%v", songList[0].SuggestedBy, sessionID)},
		}).
		Decode(&suggestingUser)
	assert.NoError(t, err)
	oldScore := suggestingUser.Score

	// make sure user is in upvoters but not in downvoters
	assert.NotEqual(t, -1, util.Find(len(upvoters),
		func(i int) bool {
			return upvoters[i] == username
		}))
	assert.Equal(t, -1, util.Find(len(downvoters),
		func(i int) bool {
			return downvoters[i] == username
		}))

	// request upvote
	resp, err := UserUpvote(username, secret, sessionID, songID)
	assert.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusOK, resp.StatusCode)

	// make sure user score was decremented
	err = userCollection.FindOne(context.Background(),
		bson.D{{"_id", fmt.Sprintf("%v@%v", songList[0].SuggestedBy, sessionID)}}).
		Decode(&suggestingUser)
	assert.NoError(t, err)
	newScore := suggestingUser.Score

	assert.Equal(t, oldScore-1, newScore)

	// make sure admin neither in downvoters nor in upvoters
	err = sessionCollection.FindOne(context.Background(),
		filter, options.FindOne().SetProjection(projection)).
		Decode(&sess)
	assert.NoError(t, err)
	songList = sess.SongList
	assert.NotEmpty(t, songList)
	upvoters = songList[0].Upvoters
	downvoters = songList[0].Downvoters

	assert.Equal(t, -1, util.Find(len(upvoters),
		func(i int) bool {
			return upvoters[i] == username
		}))
	assert.Equal(t, -1, util.Find(len(downvoters),
		func(i int) bool {
			return downvoters[i] == username
		}))
}

// Case 3: user in downvoters and not in upvoters, remove from downvoters, increment score by 2
// This test uses the Cordula Gruen, which was suggested by the admin and downvoted by the user.
// The test sends an upvote request by the user
func Test_UserUpvote_Case3(t *testing.T) {
	dropDB()
	setupDB()

	username := TestUserName
	secret := TestUserSecret
	sessionID := TestSessionID
	songID := CordulaSongID

	// get upvoters and downvoters
	filter := bson.D{
		{"_id", sessionID},
		{"song_list.id", songID},
	}
	projection := bson.D{
		{"_id", 0},
		{"song_list.$", 1},
	}
	var sess *session.Session
	err := sessionCollection.FindOne(context.Background(),
		filter,
		options.FindOne().SetProjection(projection)).
		Decode(&sess)
	assert.NoError(t, err)
	songList := sess.SongList
	assert.NotEmpty(t, songList)
	upvoters := songList[0].Upvoters
	downvoters := songList[0].Downvoters

	// get suggesting user
	var suggestingUser *user.Model
	err = userCollection.FindOne(context.Background(),
		bson.D{
			{"_id", fmt.Sprintf("%v@%v", songList[0].SuggestedBy, sessionID)},
		}).
		Decode(&suggestingUser)
	assert.NoError(t, err)
	oldScore := suggestingUser.Score

	// make sure user is in downvoters but not in upvoters
	assert.Equal(t, -1, util.Find(len(upvoters),
		func(i int) bool {
			return upvoters[i] == username
		}))
	assert.NotEqual(t, -1, util.Find(len(downvoters),
		func(i int) bool {
			return downvoters[i] == username
		}))

	// request upvote
	resp, err := UserUpvote(username, secret, sessionID, songID)
	assert.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusOK, resp.StatusCode)

	// make sure admin score was incremented
	err = userCollection.FindOne(context.Background(),
		bson.D{{"_id", fmt.Sprintf("%v@%v", songList[0].SuggestedBy, sessionID)}}).
		Decode(&suggestingUser)
	assert.NoError(t, err)
	newScore := suggestingUser.Score

	assert.Equal(t, oldScore+2, newScore)

	// make sure user is in upvoters but not in downvoters
	err = sessionCollection.FindOne(context.Background(), filter, options.FindOne().SetProjection(projection)).
		Decode(&sess)
	assert.NoError(t, err)
	songList = sess.SongList
	assert.NotEmpty(t, songList)
	upvoters = songList[0].Upvoters
	downvoters = songList[0].Downvoters

	assert.NotEqual(t, -1, util.Find(len(upvoters),
		func(i int) bool {
			return upvoters[i] == username
		}))
	assert.Equal(t, -1, util.Find(len(downvoters),
		func(i int) bool {
			return downvoters[i] == username
		}))
}

// Case 1: user neither upvoter nor downvoter: add to downvoters, decrement score by 1
// This test uses the song Skifoan, which was suggested by the admin and neither upvoted nor downvoted by the user
// The test sends a downvote request by the user
func Test_UserDownvote_Case1(t *testing.T) {
	dropDB()
	setupDB()

	username := TestUserName
	secret := TestUserSecret
	sessionID := TestSessionID
	songID := SkiFoanID

	// get upvoters and downvoters
	filter := bson.D{
		{"_id", sessionID},
		{"song_list.id", songID},
	}
	projection := bson.D{
		{"_id", 0},
		{"song_list.$", 1},
	}
	var sess *session.Session
	err := sessionCollection.FindOne(
		context.Background(),
		filter,
		options.FindOne().SetProjection(projection),
	).Decode(&sess)
	assert.NoError(t, err)
	songList := sess.SongList
	assert.NotEmpty(t, songList)
	upvoters := songList[0].Upvoters
	downvoters := songList[0].Downvoters

	// get suggesting user
	var suggestingUser *user.Model
	err = userCollection.FindOne(
		context.Background(),
		bson.D{
			{"_id", fmt.Sprintf("%v@%v", songList[0].SuggestedBy, sessionID)},
		},
	).Decode(&suggestingUser)
	assert.NoError(t, err)
	oldScore := suggestingUser.Score

	// make sure user is neither in upvoters nor in downvoters
	assert.Equal(t, -1, util.Find(len(upvoters),
		func(i int) bool {
			return upvoters[i] == username
		}))
	assert.Equal(t, -1, util.Find(len(downvoters),
		func(i int) bool {
			return downvoters[i] == username
		}))

	// request downvote
	resp, err := UserDownvote(username, secret, sessionID, songID)
	assert.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusOK, resp.StatusCode)

	// make sure admin score was decremented
	err = userCollection.FindOne(
		context.Background(),
		bson.D{
			{"_id", fmt.Sprintf("%v@%v", songList[0].SuggestedBy, sessionID)},
		},
	).Decode(&suggestingUser)
	assert.NoError(t, err)
	newScore := suggestingUser.Score

	assert.Equal(t, oldScore-1, newScore)

	// make sure user is in downvoters but not in upvoters
	err = sessionCollection.FindOne(context.Background(), filter, options.FindOne().SetProjection(projection)).
		Decode(&sess)
	assert.NoError(t, err)
	songList = sess.SongList
	assert.NotEmpty(t, songList)
	upvoters = songList[0].Upvoters
	downvoters = songList[0].Downvoters

	assert.Equal(t, -1, util.Find(len(upvoters),
		func(i int) bool {
			return upvoters[i] == username
		}))
	assert.NotEqual(t, -1, util.Find(len(downvoters),
		func(i int) bool {
			return downvoters[i] == username
		}))
}

// Case 2: admin in upvoters and not in downvoters, remove from upvoters, decrement score by 2
// This test uses the song Anton aus Tirol, which was suggested by the user and upvoted by the admin.
// The test sends a downvote request by the admin
func Test_UserDownvote_Case2(t *testing.T) {
	dropDB()
	setupDB()

	username := TestAdminUsername
	secret := TestAdminSecret
	sessionID := TestSessionID
	songID := AntonAusTirolID

	// get upvoters and downvoters
	filter := bson.D{
		{"_id", sessionID},
		{"song_list.id", songID},
	}
	projection := bson.D{
		{"_id", 0},
		{"song_list.$", 1},
	}
	var sess *session.Session
	err := sessionCollection.FindOne(context.Background(),
		filter,
		options.FindOne().SetProjection(projection)).
		Decode(&sess)
	assert.NoError(t, err)
	songList := sess.SongList
	assert.NotEmpty(t, songList)
	upvoters := songList[0].Upvoters
	downvoters := songList[0].Downvoters

	// get suggesting user
	var suggestingUser *user.Model
	err = userCollection.FindOne(context.Background(),
		bson.D{
			{"_id", fmt.Sprintf("%v@%v", songList[0].SuggestedBy, sessionID)},
		}).
		Decode(&suggestingUser)
	assert.NoError(t, err)
	oldScore := suggestingUser.Score

	// make sure user is in upvoters but not in downvoters
	assert.NotEqual(t, -1, util.Find(len(upvoters),
		func(i int) bool {
			return upvoters[i] == username
		}))
	assert.Equal(t, -1, util.Find(len(downvoters),
		func(i int) bool {
			return downvoters[i] == username
		}))

	// request downvote
	resp, err := UserDownvote(username, secret, sessionID, songID)
	assert.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusOK, resp.StatusCode)

	// make sure user score was decremented
	err = userCollection.FindOne(context.Background(),
		bson.D{{"_id", fmt.Sprintf("%v@%v", songList[0].SuggestedBy, sessionID)}}).
		Decode(&suggestingUser)
	assert.NoError(t, err)
	newScore := suggestingUser.Score

	assert.Equal(t, oldScore-2, newScore)

	// make sure admin is in downvoters but not in upvoters
	err = sessionCollection.FindOne(context.Background(),
		filter, options.FindOne().SetProjection(projection)).
		Decode(&sess)
	assert.NoError(t, err)
	songList = sess.SongList
	assert.NotEmpty(t, songList)
	upvoters = songList[0].Upvoters
	downvoters = songList[0].Downvoters

	assert.Equal(t, -1, util.Find(len(upvoters),
		func(i int) bool {
			return upvoters[i] == username
		}))
	assert.NotEqual(t, -1, util.Find(len(downvoters),
		func(i int) bool {
			return downvoters[i] == username
		}))
}

// Case 3: user in downvoters and not in upvoters, remove from downvoters, increment score by 1
// This test uses the Cordula Gruen, which was suggested by the admin and downvoted by the user.
// The test sends an upvote request by the user
func Test_UserDownvote_Case3(t *testing.T) {
	dropDB()
	setupDB()

	username := TestUserName
	secret := TestUserSecret
	sessionID := TestSessionID
	songID := CordulaSongID

	// get upvoters and downvoters
	filter := bson.D{
		{"_id", sessionID},
		{"song_list.id", songID},
	}
	projection := bson.D{
		{"_id", 0},
		{"song_list.$", 1},
	}
	var sess *session.Session
	err := sessionCollection.FindOne(context.Background(),
		filter,
		options.FindOne().SetProjection(projection)).
		Decode(&sess)
	assert.NoError(t, err)
	songList := sess.SongList
	assert.NotEmpty(t, songList)
	upvoters := songList[0].Upvoters
	downvoters := songList[0].Downvoters

	// get suggesting user
	var suggestingUser *user.Model
	err = userCollection.FindOne(context.Background(),
		bson.D{
			{"_id", fmt.Sprintf("%v@%v", songList[0].SuggestedBy, sessionID)},
		}).
		Decode(&suggestingUser)
	assert.NoError(t, err)
	oldScore := suggestingUser.Score

	// make sure user is in downvoters but not in upvoters
	assert.Equal(t, -1, util.Find(len(upvoters),
		func(i int) bool {
			return upvoters[i] == username
		}))
	assert.NotEqual(t, -1, util.Find(len(downvoters),
		func(i int) bool {
			return downvoters[i] == username
		}))

	// request downvote
	resp, err := UserDownvote(username, secret, sessionID, songID)
	assert.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusOK, resp.StatusCode)

	// make sure admin score was incremented
	err = userCollection.FindOne(context.Background(),
		bson.D{{"_id", fmt.Sprintf("%v@%v", songList[0].SuggestedBy, sessionID)}}).
		Decode(&suggestingUser)
	assert.NoError(t, err)
	newScore := suggestingUser.Score

	assert.Equal(t, oldScore+1, newScore)

	// make sure user is neither in upvoters nor in downvoters
	err = sessionCollection.FindOne(context.Background(), filter, options.FindOne().SetProjection(projection)).
		Decode(&sess)
	assert.NoError(t, err)
	songList = sess.SongList
	assert.NotEmpty(t, songList)
	upvoters = songList[0].Upvoters
	downvoters = songList[0].Downvoters

	assert.Equal(t, -1, util.Find(len(upvoters),
		func(i int) bool {
			return upvoters[i] == username
		}))
	assert.Equal(t, -1, util.Find(len(downvoters),
		func(i int) bool {
			return downvoters[i] == username
		}))
}

// admin removes the song Skifoan, which was suggested by himself
func Test_AdminRemoveSong_AdminSong(t *testing.T) {
	dropDB()
	setupDB()

	username := TestAdminUsername
	secret := TestAdminSecret
	sessionID := TestSessionID
	songID := SkiFoanID

	// get song from db before request
	filter := bson.D{
		{"_id", sessionID},
		{"song_list.id", songID},
	}
	projection := bson.D{
		{"_id", 0},
		{"song_list.$", 1},
	}
	var sess *session.Session
	err := sessionCollection.FindOne(
		context.Background(),
		filter,
		options.FindOne().SetProjection(projection),
	).Decode(&sess)
	assert.NoError(t, err)
	songList := sess.SongList
	assert.NotEmpty(t, songList)

	resp, err := AdminRemoveSong(username, secret, sessionID, songID)
	assert.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	// make sure song was deleted
	err = sessionCollection.FindOne(
		context.Background(),
		filter,
		options.FindOne().SetProjection(projection),
	).Decode(&sess)
	assert.Error(t, err) // no document in result
}

// admin removes the song Anton aus Tirol, which was suggested by the user
func Test_AdminRemoveSong_UserSong(t *testing.T) {
	dropDB()
	setupDB()

	username := TestAdminUsername
	secret := TestAdminSecret
	sessionID := TestSessionID
	songID := AntonAusTirolID

	// get song from db before request
	filter := bson.D{
		{"_id", sessionID},
		{"song_list.id", songID},
	}
	projection := bson.D{
		{"_id", 0},
		{"song_list.$", 1},
	}
	var sess *session.Session
	err := sessionCollection.FindOne(
		context.Background(),
		filter,
		options.FindOne().SetProjection(projection),
	).Decode(&sess)
	assert.NoError(t, err)
	songList := sess.SongList
	assert.NotEmpty(t, songList)

	resp, err := AdminRemoveSong(username, secret, sessionID, songID)
	assert.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	// make sure song was deleted
	err = sessionCollection.FindOne(
		context.Background(),
		filter,
		options.FindOne().SetProjection(projection),
	).Decode(&sess)
	assert.Error(t, err) // no documents in collection
}

// user attemps to remove the song Skifoan, which was suggested by the admin
// This test also implicitly tests the admin auth function
func Test_AdminRemoveSong_UserRequest_AdminSong(t *testing.T) {
	dropDB()
	setupDB()

	username := TestUserName
	secret := TestUserSecret
	sessionID := TestSessionID
	songID := SkiFoanID

	// get song from db before request
	filter := bson.D{
		{"_id", sessionID},
		{"song_list.id", songID},
	}
	projection := bson.D{
		{"_id", 0},
		{"song_list.$", 1},
	}
	var sess *session.Session
	err := sessionCollection.FindOne(
		context.Background(),
		filter,
		options.FindOne().SetProjection(projection),
	).Decode(&sess)
	assert.NoError(t, err)
	songList := sess.SongList
	assert.NotEmpty(t, songList)

	resp, err := AdminRemoveSong(username, secret, sessionID, songID)
	assert.NoError(t, err)
	defer resp.Body.Close()
	assert.NotEqual(t, http.StatusOK, resp.StatusCode) // not authorized

	// make sure song was not deleted
	err = sessionCollection.FindOne(
		context.Background(),
		filter,
		options.FindOne().SetProjection(projection),
	).Decode(&sess)
	assert.NoError(t, err)
	songList = sess.SongList
	assert.NotEmpty(t, songList)
}

// user attempts to remove the song Anton aus Tirol, which was suggested by himself
func Test_AdminRemoveSong_UserRequest_UserSong(t *testing.T) {
	dropDB()
	setupDB()

	username := TestUserName
	secret := TestUserSecret
	sessionID := TestSessionID
	songID := AntonAusTirolID

	// get song from db before request
	filter := bson.D{
		{"_id", sessionID},
		{"song_list.id", songID},
	}
	projection := bson.D{
		{"_id", 0},
		{"song_list.$", 1},
	}
	var sess *session.Session
	err := sessionCollection.FindOne(
		context.Background(),
		filter,
		options.FindOne().SetProjection(projection),
	).Decode(&sess)
	assert.NoError(t, err)
	songList := sess.SongList
	assert.NotEmpty(t, songList)

	resp, err := AdminRemoveSong(username, secret, sessionID, songID)
	assert.NoError(t, err)
	defer resp.Body.Close()
	assert.NotEqual(t, http.StatusOK, resp.StatusCode) // not authorized

	// make sure song was not deleted
	err = sessionCollection.FindOne(
		context.Background(),
		filter,
		options.FindOne().SetProjection(projection),
	).Decode(&sess)
	assert.NoError(t, err)
	songList = sess.SongList
	assert.NotEmpty(t, songList)
}
