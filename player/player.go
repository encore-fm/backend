package player

import "time"

// every session object in the db contains a player object
type Player struct {
	SongID       string        `json:"current_song_id" bson:"current_song_id"`
	SongProgress time.Duration `json:"progress" bson:"progress"`
	SongEnd      time.Time     `json:"song_end" bson:"song_end"`
	Paused       bool          `json:"paused" bson:"paused"`
}
