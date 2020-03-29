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
	assert.Equal(t, float64(1), result.Score)
	assert.False(t, result.IsAdmin)
	assert.Equal(t, 128, len(result.Secret))
}

func TestNewAdmin(t *testing.T) {
	username := "test"
	sessionID := "session_id"
	result, err := NewAdmin(username, sessionID)

	assert.Nil(t, err)
	assert.Equal(t, username, result.Username)
	assert.Equal(t, sessionID, result.SessionID)

	assert.Equal(t, float64(1), result.Score)
	assert.True(t, result.IsAdmin)
	assert.Equal(t, 128, len(result.Secret))
}
