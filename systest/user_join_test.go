// +build !ci

package systest

import (
	"context"
	"encoding/json"
	"net/http"
	"testing"

	"github.com/encore-fm/backend/handlers"
	"github.com/encore-fm/backend/user"
	"github.com/encore-fm/backend/util"
	"github.com/stretchr/testify/assert"
	"go.mongodb.org/mongo-driver/bson"
)

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
	var response struct {
		UserInfo *user.Model `json:"user_info"`
		AuthUrl  string      `json:"auth_url"`
	}
	err = json.NewDecoder(resp.Body).Decode(&response)
	assert.NoError(t, err)

	// make sure username und session id match, user is not admin and score is initialized with 1
	assert.Equal(t, username, response.UserInfo.Username)
	assert.Equal(t, TestSessionID, response.UserInfo.SessionID)
	assert.Equal(t, 1, response.UserInfo.Score)
	assert.Equal(t, false, response.UserInfo.IsAdmin)

	// make sure data is written into db
	var foundUser *user.Model
	err = userCollection.
		FindOne(context.Background(), bson.D{{"_id", response.UserInfo.ID}}).
		Decode(&foundUser)
	assert.NoError(t, err)

	// set fields that are not in response to nil
	foundUser.AuthState = ""
	foundUser.AuthToken = nil
	assert.Equal(t, response.UserInfo, foundUser)

	// get new count
	newCount, err := userCollection.CountDocuments(context.Background(), bson.D{})
	assert.NoError(t, err)
	// make sure the db only added one document to usercollection
	assert.Equal(t, count+1, newCount)
}

func Test_UserJoin_NonExistingSession(t *testing.T) {
	username := "eti"
	sessionID, err := util.GenerateSecret(16)
	assert.NoError(t, err)

	// get db collection count before insertion
	count, err := userCollection.CountDocuments(context.Background(), bson.D{})
	assert.NoError(t, err)

	// post request, expect http status not OK
	resp, err := UserJoin(username, sessionID)
	assert.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, http.StatusBadRequest, resp.StatusCode) // expect 404 when session is not found

	// deserialize response body and assert expected results
	var response handlers.FrontendError

	err = json.NewDecoder(resp.Body).Decode(&response)
	assert.NoError(t, err)

	// make sure the correct frontenderror is returned
	assert.Equal(t, handlers.SessionNotFoundError, response)

	// get new count
	newCount, err := userCollection.CountDocuments(context.Background(), bson.D{})
	assert.NoError(t, err)
	// make sure no new documents were added to usercollection
	assert.Equal(t, count, newCount)
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
	// 409 expected when username exists
	assert.Equal(t, http.StatusConflict, resp.StatusCode)

	// deserialize response body and assert expected results
	var response handlers.FrontendError

	err = json.NewDecoder(resp.Body).Decode(&response)
	assert.NoError(t, err)
	assert.Equal(t, handlers.UserConflictError, response)

	// get new count
	newCount, err := userCollection.CountDocuments(context.Background(), bson.D{})
	assert.NoError(t, err)
	// make sure no new documents were added to usercollection
	assert.Equal(t, count, newCount)
}
