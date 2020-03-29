package user

import (
	"github.com/antonbaumann/spotify-jukebox/util"
	"golang.org/x/oauth2"
)

type Model struct {
	Username          string  `json:"username" bson:"_id"`
	Secret            string  `json:"secret" bson:"secret"`
	SessionID         string  `json:"session_id" bson:"session_id"`
	IsAdmin           bool    `json:"is_admin" bson:"is_admin"`
	Score             float64 `json:"score" bson:"score"`
	SpotifyAuthorized bool    `json:"spotify_authorized" bson:"spotify_authorized"`

	AuthToken *oauth2.Token `json:"auth_token" bson:"auth_token"`
	AuthState string        `json:"auth_state" bson:"auth_state"`
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
		Username:          username,
		Secret:            secret,
		SessionID:         sessionID,
		IsAdmin:           false,
		Score:             1,
		AuthState:         state,
		SpotifyAuthorized: false,
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
