package song

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/zmb3/spotify"
)

func TestNew(t *testing.T) {
	username := "test"
	songScore := float64(42)
	songID := "song_id"
	songName := "song_name"
	previewURL := "preview_url"
	duration := 1000
	artistStrings := []string{"artist_1", "artist_2", "artist_3"}

	info := &spotify.FullTrack{
		SimpleTrack: spotify.SimpleTrack{
			Artists: []spotify.SimpleArtist{
				{Name: "artist_1"},
				{Name: "artist_2"},
				{Name: "artist_3"},
			},
			Duration:   duration,
			ID:         spotify.ID(songID),
			Name:       songName,
			PreviewURL: previewURL,
		},
	}

	result := New(username, songScore, info)
	assert.Equal(t, songName, result.Name)
	assert.Equal(t, previewURL, result.PreviewUrl)
	assert.Equal(t, songID, result.ID)
	assert.Equal(t, duration, result.Duration)
	assert.Equal(t, artistStrings, result.Artists)
}
