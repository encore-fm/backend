// +build !ci

package systest

import (
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_UserPing(t *testing.T) {
	username := TestAdminUsername
	secret := TestAdminSecret
	sessionID := TestSessionID

	resp, err := UserPing(username, secret, sessionID)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func Test_UserPing_WrongCredentials(t *testing.T) {
	username := TestAdminUsername
	secret := "wrong secret"
	sessionID := TestSessionID

	resp, err := UserPing(username, secret, sessionID)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusUnauthorized, resp.StatusCode)
}
