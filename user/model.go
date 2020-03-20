package user

import (
	"crypto/rand"
	"fmt"
)

type Model struct {
	Username string  `json:"username" bson:"username"`
	Secret   string  `json:"secret" bson:"secret"`
	Score    float64 `json:"score" bson:"score"`
}

// returns a 128 char secret key
func GenerateSecret() (string, error) {
	key := make([]byte, 64)
	_, err := rand.Read(key)
	if err != nil {
		return "", fmt.Errorf("generate secret: %v", err)
	}
	return fmt.Sprintf("%x", key), nil
}

func New(username string) (*Model, error) {
	secret, err := GenerateSecret()
	if err != nil {
		return nil, err
	}
	model := &Model{
		Username: username,
		Secret:   secret,
		Score:    0,
	}
	return model, nil
}