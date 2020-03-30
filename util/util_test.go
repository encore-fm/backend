package util

import (
	"errors"
	"regexp"
	"testing"
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
