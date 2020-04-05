package systest

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"golang.org/x/oauth2"
)

func Test_UserGetClientToken(t *testing.T) {
	dropDB()
	setupDB()

	resp, err := UserGetClientToken(TestAdminUsername, TestAdminSecret, TestSessionID)
	assert.NoError(t, err)

	var token *oauth2.Token
	err = json.NewDecoder(resp.Body).Decode(&token)
	assert.NoError(t, err)

	assert.NotEmpty(t, token.AccessToken)
	assert.NotEmpty(t, token.Expiry)
}
