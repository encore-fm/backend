package playerctrl

import (
	"time"

	"github.com/antonbaumann/spotify-jukebox/events"
)

// define song added event
const SongAdded events.EventType = "player_event:song_added"

// define play / paused event
const PlayPauseEvent events.EventType = "player_event:play_pause"

type PlayPausePayload struct {
	Paused bool `json:"paused"`
}

// define skip event
const SkipEvent events.EventType = "player_event:skip"

type SkipPayload struct{}

// define seek event
const SeekEvent events.EventType = "player_event:seek"

type SeekPayload struct {
	Progress time.Duration `json:"progress"`
}

// define set synchronized event
const SetSynchronized events.EventType = "player_event:set_synchronized"

type SetSynchronizedPayload struct {
	UserID       string
	Synchronized bool
}

// define set synchronized event
const Synchronize events.EventType = "player_event:synchronize"

type SynchronizePayload struct {
	UserID string
}

// define reset event
const ResetEvent events.EventType = "player_event:reset_session"

type ResetPayload struct {
	SessionID string `json:"session_id"`
}
