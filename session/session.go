package session

import "github.com/antonbaumann/spotify-jukebox/util"

// Session stores session information
// todo: maybe save session options later
type Session struct {
	ID string `json:"id" bson:"_id"`
}

func New() (*Session, error) {
	sessionID, err := util.GenerateSecret()
	if err != nil {
		return nil, err
	}
	return &Session{ID: sessionID}, nil
}
