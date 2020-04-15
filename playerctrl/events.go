package playerctrl

import "github.com/antonbaumann/spotify-jukebox/events"

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

// define reset event
const ResetEvent events.EventType = "player_event:reset_session"

type ResetPayload struct {
	SessionID string `json:"session_id"`
}
