package user

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNew(t *testing.T) {
	username := "test"
	sessionID := "session_id"
	result, err := New(username, sessionID)

	assert.Nil(t, err)
	assert.Equal(t, username, result.Username)
	assert.Equal(t, sessionID, result.SessionID)
	assert.Equal(t, 1, result.Score)
	assert.False(t, result.IsAdmin)
	assert.False(t, result.SpotifyAuthorized)
	assert.Equal(t, 128, len(result.Secret))
}

func TestNewAdmin(t *testing.T) {
	username := "test"
	sessionID := "session_id"
	result, err := NewAdmin(username, sessionID)

	assert.Nil(t, err)
	assert.Equal(t, username, result.Username)
	assert.Equal(t, sessionID, result.SessionID)

	assert.Equal(t, 1, result.Score)
	assert.True(t, result.IsAdmin)
	assert.False(t, result.SpotifyAuthorized)
	assert.Equal(t, 128, len(result.Secret))
}

func TestNew_InvalidUsername(t *testing.T) {
	username := "12"
	_, err := New(username, "")
	assert.Equal(t, ErrUsernameTooShort, err)

	username = "aaaaaaaaaaaaaaaaaaaaa" // 21 * a
	_, err = New(username, "")
	assert.Equal(t, ErrUsernameTooLong, err)

	username = "test!user~"
	_, err = New(username, "")
	assert.Equal(t, ErrUsernameInvalidCharacter, err)
}
