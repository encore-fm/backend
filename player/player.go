package player

import (
	"time"

	"github.com/antonbaumann/spotify-jukebox/song"
)

// every session object in the db contains a player object
type Player struct {
	CurrentSong   *song.Model   `json:"current_song" bson:"current_song"`
	SongStart     time.Time     `json:"song_start" bson:"song_start"`
	PauseStart    time.Time     `json:"pause_start" bson:"pause_start"`
	PauseDuration time.Duration `json:"pause_duration" bson:"pause_duration"`
	Paused        bool          `json:"paused" bson:"paused"`
}

func (p *Player) Progress() time.Duration {
	return time.Now().Sub(p.SongStart) - p.PauseDuration
}
