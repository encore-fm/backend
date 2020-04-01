package user

import (
	"fmt"

	"github.com/antonbaumann/spotify-jukebox/util"
	"golang.org/x/oauth2"
)

type Score int

type Model struct {
	ID                string `json:"id" bson:"_id"`
	Username          string `json:"username" bson:"username"`
	Secret            string `json:"secret" bson:"secret"`
	SessionID         string `json:"session_id" bson:"session_id"`
	IsAdmin           bool   `json:"is_admin" bson:"is_admin"`
	Score             Score  `json:"score" bson:"score"`
	SpotifyAuthorized bool   `json:"spotify_authorized" bson:"spotify_authorized"`

	AuthToken *oauth2.Token `json:"auth_token" bson:"auth_token"`
	AuthState string        `json:"auth_state" bson:"auth_state"`
}

type ListElement struct {
	Username string `json:"username" bson:"username"`
	IsAdmin  bool   `json:"is_admin" bson:"is_admin"`
	Score    Score  `json:"score" bson:"score"`
}

func GenerateUserID(username, userID string) string {
	return fmt.Sprintf("%v@%v", username, userID)
}

func New(username, sessionID string) (*Model, error) {
	secret, err := util.GenerateSecret()
	if err != nil {
		return nil, err
	}

	state, err := util.GenerateSecret()
	if err != nil {
		return nil, err
	}

	model := &Model{
		ID:                GenerateUserID(username, sessionID),
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

func NewAdmin(username, sessionID string) (*Model, error) {
	admin, err := New(username, sessionID)
	if err != nil {
		return nil, err
	}
	admin.IsAdmin = true
	return admin, nil
}
