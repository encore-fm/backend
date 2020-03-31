package song

import (
	"testing"
	"time"

	"github.com/antonbaumann/spotify-jukebox/user"
	"github.com/stretchr/testify/assert"
	"github.com/zmb3/spotify"
)

func TestNew1(t *testing.T) {
	username := "username"
	songScore := float64(42)

	expected := &Model{
		ID:          "song_id",
		Name:        "song_name",
		Artists:     []string{"artist_1", "artist_2", "artist_3"},
		Duration:    100,
		CoverUrl:    "cover_url",
		AlbumName:   "album_name",
		PreviewUrl:  "preview_url",
		SuggestedBy: username,
		Score:       songScore,
		TimeAdded:   time.Time{},
		Upvoters:    user.NewVoters(),
		Downvoters:  user.NewVoters(),
	}
	info := &spotify.FullTrack{
		SimpleTrack: spotify.SimpleTrack{
			Artists: []spotify.SimpleArtist{
				{Name: "artist_1"},
				{Name: "artist_2"},
				{Name: "artist_3"},
			},
			Duration:   expected.Duration,
			ID:         spotify.ID(expected.ID),
			Name:       expected.Name,
			PreviewURL: expected.PreviewUrl,
		},
		Album: spotify.SimpleAlbum{
			Name: expected.AlbumName,
			Images: []spotify.Image{
				{URL: expected.CoverUrl},
			},
		},
	}

	result := New(username, songScore, info)

	//overwrite result.TimeAdded
	result.TimeAdded = time.Time{}

	assert.Equal(t, expected, result)
}

func TestNew2(t *testing.T) {
	username := "username"
	songScore := float64(42)

	expected := &Model{
		ID:          "song_id",
		Name:        "song_name",
		Artists:     []string{"artist_1", "artist_2", "artist_3"},
		Duration:    100,
		CoverUrl:    "",
		AlbumName:   "album_name",
		PreviewUrl:  "preview_url",
		SuggestedBy: username,
		Score:       songScore,
		TimeAdded:   time.Time{},
		Upvoters:    user.NewVoters(),
		Downvoters:  user.NewVoters(),
	}
	info := &spotify.FullTrack{
		SimpleTrack: spotify.SimpleTrack{
			Artists: []spotify.SimpleArtist{
				{Name: "artist_1"},
				{Name: "artist_2"},
				{Name: "artist_3"},
			},
			Duration:   expected.Duration,
			ID:         spotify.ID(expected.ID),
			Name:       expected.Name,
			PreviewURL: expected.PreviewUrl,
		},
		Album: spotify.SimpleAlbum{
			Name:   expected.AlbumName,
			Images: []spotify.Image{},
		},
	}

	result := New(username, songScore, info)

	//overwrite result.TimeAdded
	result.TimeAdded = time.Time{}

	assert.Equal(t, expected, result)
}
