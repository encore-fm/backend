package user

import (
	"errors"
	"regexp"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGenerateSecret(t *testing.T) {
	alphanumRegex := regexp.MustCompile("^[a-zA-Z0-9]{128}$")
	secret, err := GenerateSecret()
	if err != nil {
		t.Error(err)
	}
	if !alphanumRegex.MatchString(secret) {
		t.Error(errors.New("secret must have len 128 and only contain alphanumeric characters"))
	}
}

func TestNew(t *testing.T) {
	username := "test"
	result, err := New(username)

	assert.Nil(t, err)
	assert.Equal(t, username, result.Username)
	assert.Equal(t, float64(0), result.Score)
	assert.False(t, result.IsAdmin)
	assert.Equal(t, 128, len(result.Secret))
}

func TestNewAdmin(t *testing.T) {
	username := "test"
	result, err := NewAdmin(username)

	assert.Nil(t, err)
	assert.Equal(t, username, result.Username)
	assert.Equal(t, float64(0), result.Score)
	assert.True(t, result.IsAdmin)
	assert.Equal(t, 128, len(result.Secret))
}
