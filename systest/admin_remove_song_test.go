// +build !ci

package systest

import (
	"context"
	"net/http"
	"testing"

	"github.com/encore-fm/backend/session"
	"github.com/stretchr/testify/assert"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo/options"
)

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
