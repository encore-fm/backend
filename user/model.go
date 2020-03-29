package user

import (
	"github.com/antonbaumann/spotify-jukebox/util"
)

type Model struct {
	Username  string  `json:"username" bson:"_id"`
	Secret    string  `json:"secret" bson:"secret"`
	SessionID string  `json:"session_id" bson:"session_id"`
	IsAdmin   bool    `json:"is_admin" bson:"is_admin"`
	Score     float64 `json:"score" bson:"score"`

	AuthState string `json:"state" bson:"state"`
}

type ListElement struct {
	Username string  `json:"username" bson:"_id"`
	IsAdmin  bool    `json:"is_admin" bson:"is_admin"`
	Score    float64 `json:"score" bson:"score"`
}

func New(username string, sessionID string) (*Model, error) {
	secret, err := util.GenerateSecret()
	if err != nil {
		return nil, err
	}

	state, err := util.GenerateSecret()
	if err != nil {
		return nil, err
	}

	model := &Model{
		Username:  username,
		Secret:    secret,
		SessionID: sessionID,
		IsAdmin:   false,
		Score:     1,
		AuthState: state,
	}
	return model, nil
}

func NewAdmin(username string, sessionID string) (*Model, error) {
	admin, err := New(username, sessionID)
	if err != nil {
		return nil, err
	}
	admin.IsAdmin = true
	return admin, nil
}
