package player

import (
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestPlayer_Progress_Playing(t *testing.T) {
	now := time.Now()
	player := &Player{
		CurrentSong:   nil,
		SongStart:     now.Add(-5 * time.Minute),
		PauseStart:    time.Time{},
		PauseDuration: 1 * time.Minute,
		Paused:        false,
	}

	assert.WithinDuration(t, now.Add(player.Progress()), now.Add(4*time.Minute), 200*time.Millisecond)
}

func TestPlayer_Progress_Paused(t *testing.T) {
	now := time.Now()
	player := &Player{
		CurrentSong:   nil,
		SongStart:     now.Add(-5 * time.Minute),
		PauseStart:    now.Add(-3 * time.Minute),
		PauseDuration: 1 * time.Minute,
		Paused:        true,
	}

	fmt.Println(player.Progress())
	assert.WithinDuration(t, now.Add(player.Progress()), now.Add(1*time.Minute), 200*time.Millisecond)
}
