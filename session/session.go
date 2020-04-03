package session

import (
	"github.com/antonbaumann/spotify-jukebox/song"
	"github.com/antonbaumann/spotify-jukebox/util"
)

// Session stores session information
// todo: maybe save session options later
type Session struct {
	ID       string        `json:"id" bson:"_id"`
	SongList []*song.Model `json:"song_list" bson:"song_list"`
}

func New() (*Session, error) {
	sessionID, err := util.GenerateSecret()
	if err != nil {
		return nil, err
	}
	return &Session{
		ID:       sessionID,
		SongList: make([]*song.Model, 0),
	}, nil
}
