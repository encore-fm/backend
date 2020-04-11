package util

import (
	"crypto/rand"
	"fmt"
)

// returns a 128 char secret key
func GenerateSecret(bytes int) (string, error) {
	key := make([]byte, bytes)
	_, err := rand.Read(key)
	if err != nil {
		return "", fmt.Errorf("generate secret: %v", err)
	}
	return fmt.Sprintf("%x", key), nil
}

func Find(limit int, predicate func(i int) bool) int {
	for i := 0; i < limit; i++ {
		if predicate(i) {
			return i
		}
	}
	return -1
}
