package sse

import (
	"time"

	"github.com/antonbaumann/spotify-jukebox/events"
	"github.com/antonbaumann/spotify-jukebox/song"
)

const (
	PlaylistChange    events.EventType = "sse:playlist_change"
	PlayerStateChange events.EventType = "sse:player_state_change"
	UserListChange    events.EventType = "sse:user_list_change"
)

type PlayerStateChangePayload struct {
	CurrentSong *song.Model `json:"current_song"`
	IsPlaying   bool        `json:"is_playing"`
	ProgressMs  int64       `json:"progress"`
	Timestamp   time.Time   `json:"timestamp"`
}
