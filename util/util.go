package util

import (
	"crypto/rand"
	"fmt"
)

// returns a 128 char secret key
func GenerateSecret() (string, error) {
	key := make([]byte, 64)
	_, err := rand.Read(key)
	if err != nil {
		return "", fmt.Errorf("generate secret: %v", err)
	}
	return fmt.Sprintf("%x", key), nil
}