package util

import (
	"errors"
	"github.com/magiconair/properties/assert"
	"regexp"
	"testing"
)

func TestGenerateSecret(t *testing.T) {
	alphanumRegex := regexp.MustCompile("^[a-zA-Z0-9]{128}$")
	secret, err := GenerateSecret(64)
	if err != nil {
		t.Error(err)
	}
	if !alphanumRegex.MatchString(secret) {
		t.Error(errors.New("secret must have len 128 and only contain alphanumeric characters"))
	}
}

func TestFind(t *testing.T) {
	strings := []string{"haystack1", "haystack2", "haystack3"}
	needle := "haystack2"

	result := Find(len(strings), func(i int) bool {
		return strings[i] == needle
	})

	assert.Equal(t, 1, result)
}
