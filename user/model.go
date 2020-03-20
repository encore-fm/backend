package user

import (
	"crypto/rand"
	"fmt"
)

type Model struct {
	Username string  `json:"username" bson:"username"`
	Secret   string  `json:"secret" bson:"secret"`
	IsAdmin  bool    `json:"is_admin" bson:"is_admin"`
	Score    float64 `json:"score" bson:"score"`
}

type ListElement struct {
	Username string  `json:"username" bson:"username"`
	IsAdmin  bool    `json:"is_admin" bson:"is_admin"`
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
		IsAdmin:  false,
		Score:    0,
	}
	return model, nil
}

func NewAdmin(username string) (*Model, error) {
	secret, err := GenerateSecret()
	if err != nil {
		return nil, err
	}
	model := &Model{
		Username: username,
		Secret:   secret,
		IsAdmin:  true,
		Score:    0,
	}
	return model, nil
}
