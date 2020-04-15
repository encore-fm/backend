package playerctrl

import (
	"time"

	"github.com/antonbaumann/spotify-jukebox/events"
)

// define register session event
const RegisterSessionEvent events.EventType = "player_event:register_session"

type RegisterSessionPayload struct {
	SessionID string `json:"session_id"`
}

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

// define reset event
const ResetEvent events.EventType = "player_event:reset_session"

type ResetPayload struct {
	SessionID string `json:"session_id"`
}