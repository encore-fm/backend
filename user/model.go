package user

import (
	"errors"
	"fmt"
	"regexp"

	"github.com/antonbaumann/spotify-jukebox/util"
	"golang.org/x/oauth2"
)

const (
	StateBytes  = 64
	SecretBytes = 64
)

var ErrUsernameTooShort = fmt.Errorf("username must have at least %v characters", MinLen)
var ErrUsernameTooLong = fmt.Errorf("username can not have more than %v characters", MaxLen)
var ErrUsernameInvalidCharacter = errors.New("username contains invalid character")

const (
	MinLen = 3
	MaxLen = 20
)

var (
	allowedCharactersRegex = regexp.MustCompile("^[\\w.-]*$")
)

type Model struct {
	ID                string `json:"id" bson:"_id"`
	Username          string `json:"username" bson:"username"`
	Secret            string `json:"secret" bson:"secret"`
	SessionID         string `json:"session_id" bson:"session_id"`
	IsAdmin           bool   `json:"is_admin" bson:"is_admin"`
	Score             int    `json:"score" bson:"score"`
	SpotifyAuthorized bool   `json:"spotify_authorized" bson:"spotify_authorized"`

	SpotifySynchronized bool `json:"spotify_synchronized" bson:"spotify_synchronized"`

	AuthToken *oauth2.Token `json:"-" bson:"auth_token"`
	AuthState string        `json:"-" bson:"auth_state"`

	ActiveSSEConnections int `json:"-" bson:"active_sse_connections"`
}

type ListElement struct {
	Username            string `json:"username" bson:"username"`
	IsAdmin             bool   `json:"is_admin" bson:"is_admin"`
	Score               int    `json:"score" bson:"score"`
	SpotifySynchronized bool   `json:"spotify_synchronized" bson:"spotify_synchronized"`
}

type SpotifyClient struct {
	ID        string        `bson:"_id"`
	Username  string        `bson:"username"`
	SessionID string        `bson:"session_id"`
	IsAdmin   bool          `bson:"is_admin"`
	AuthToken *oauth2.Token `bson:"auth_token"`
}

func GenerateUserID(username, sessionID string) string {
	return fmt.Sprintf("%v@%v", username, sessionID)
}

func validateUsername(username string) error {
	if len(username) < MinLen {
		return ErrUsernameTooShort
	}

	if len(username) > MaxLen {
		return ErrUsernameTooLong
	}

	if !allowedCharactersRegex.MatchString(username) {
		return ErrUsernameInvalidCharacter
	}

	return nil
}

func New(username, sessionID string) (*Model, error) {
	if err := validateUsername(username); err != nil {
		return nil, err
	}

	secret, err := util.GenerateSecret(SecretBytes)
	if err != nil {
		return nil, err
	}

	state, err := util.GenerateSecret(StateBytes)
	if err != nil {
		return nil, err
	}

	model := &Model{
		ID:                   GenerateUserID(username, sessionID),
		Username:             username,
		Secret:               secret,
		SessionID:            sessionID,
		IsAdmin:              false,
		Score:                1,
		AuthState:            state,
		SpotifyAuthorized:    false,
		SpotifySynchronized:  false,
		ActiveSSEConnections: 0,
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
