package player

import (
	"github.com/antonbaumann/spotify-jukebox/song"
	"time"
)

// every session object in the db contains a player object
type Player struct {
	CurrentSong  *song.Model   `json:"current_song" bson:"current_song"`
	SongProgress time.Duration `json:"progress" bson:"progress"`
	SongStart    time.Time     `json:"song_start" bson:"song_start"`
	Paused       bool          `json:"paused" bson:"paused"`
}
