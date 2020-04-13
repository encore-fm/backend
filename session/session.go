package session

import (
	"github.com/antonbaumann/spotify-jukebox/player"
	"github.com/antonbaumann/spotify-jukebox/song"
	"github.com/antonbaumann/spotify-jukebox/util"
)

const IDBytes = 16

// Session stores session information
// todo: maybe save session options later
type Session struct {
	ID       string         `json:"id" bson:"_id"`
	SongList []*song.Model  `json:"song_list" bson:"song_list"`
	Player   *player.Player `json:"player" bson:"player"`
}

func New() (*Session, error) {
	sessionID, err := util.GenerateSecret(IDBytes)
	if err != nil {
		return nil, err
	}
	return &Session{
		ID:       sessionID,
		SongList: make([]*song.Model, 0),
	}, nil
}
