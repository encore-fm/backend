package playerctrl

import "github.com/antonbaumann/spotify-jukebox/events"

// define register session event
const RegisterSessionEvent events.EventType = "register_session"

type RegisterSessionPayload struct {
	SessionID string `json:"session_id"`
}

// define play / paused event
const PlayPauseEvent events.EventType = "play_pause"

type PlayPausePayload struct {
	Paused bool `json:"paused"`
}
