package user

import (
	"github.com/antonbaumann/spotify-jukebox/util"
)

type Model struct {
	Username string  `json:"username" bson:"_id"`
	Secret   string  `json:"secret" bson:"secret"`
	IsAdmin  bool    `json:"is_admin" bson:"is_admin"`
	Score    float64 `json:"score" bson:"score"`
}

type ListElement struct {
	Username string  `json:"username" bson:"_id"`
	IsAdmin  bool    `json:"is_admin" bson:"is_admin"`
	Score    float64 `json:"score" bson:"score"`
}

func New(username string) (*Model, error) {
	secret, err := util.GenerateSecret()
	if err != nil {
		return nil, err
	}
	model := &Model{
		Username: username,
		Secret:   secret,
		IsAdmin:  false,
		Score:    1,
	}
	return model, nil
}

func NewAdmin(username string) (*Model, error) {
	secret, err := util.GenerateSecret()
	if err != nil {
		return nil, err
	}
	model := &Model{
		Username: username,
		Secret:   secret,
		IsAdmin:  true,
		Score:    1,
	}
	return model, nil
}
