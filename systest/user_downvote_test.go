// +build !ci

package systest

import (
	"context"
	"fmt"
	"net/http"
	"testing"

	"github.com/antonbaumann/spotify-jukebox/session"
	"github.com/antonbaumann/spotify-jukebox/user"
	"github.com/antonbaumann/spotify-jukebox/util"
	"github.com/stretchr/testify/assert"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo/options"
)

//Case 1: user neither upvoter nor downvoter: add to downvoters, decrement score by 1
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
