// +build !ci

package systest

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/antonbaumann/spotify-jukebox/handlers"
	"github.com/antonbaumann/spotify-jukebox/user"
	"github.com/stretchr/testify/assert"
	"go.mongodb.org/mongo-driver/bson"
	"golang.org/x/oauth2"
)

func Test_UserGetAuthToken_Authorized(t *testing.T) {
	dropDB()
	setupDB()

	var usr user.Model
	err := userCollection.FindOne(
		context.Background(),
		bson.D{{"_id", testAdmin.ID}}).
		Decode(&usr)
	// refresh token not expected
	usr.AuthToken.RefreshToken = ""

	resp, err := UserGetAuthToken(TestAdminUsername, TestAdminSecret, TestSessionID)
	assert.NoError(t, err)

	var token *oauth2.Token
	err = json.NewDecoder(resp.Body).Decode(&token)
	assert.NoError(t, err)
	assert.Equal(t, usr.AuthToken, token)

	assert.Empty(t, token.RefreshToken)
	assert.NotEmpty(t, token.AccessToken)
}

func Test_UserGetAuthToken_NotAuthorized(t *testing.T) {
	dropDB()
	setupDB()

	resp, err := UserGetAuthToken(TestUserName, TestUserSecret, TestSessionID)
	assert.NoError(t, err)

	var response handlers.FrontendError
	err = json.NewDecoder(resp.Body).Decode(&response)
	assert.NoError(t, err)

	assert.Equal(t, handlers.SpotifyNotAuthenticatedError, response)
}
