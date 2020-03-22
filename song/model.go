package song

import (
	"time"

	"github.com/zmb3/spotify"
)

type Model struct {
	ID          string    `json:"id" bson:"id"`
	Name        string    `json:"name" bson:"name"`
	Artists     []string  `json:"artists" bson:"artists"`
	Duration    int       `json:"duration_ms" bson:"duration_ms"`
	PreviewUrl  string    `json:"preview_url" bson:"preview_url"`
	SuggestedBy string    `json:"suggested_by" bson:"suggested_by"`
	Score       float64   `json:"score" bson:"score"`
	TimeAdded   time.Time `json:"time_added" bson:"time_added"`
}

func New(
	suggestingUser string,
	score float64,
	info *spotify.FullTrack,
) *Model {
	return &Model{
		ID:          string(info.ID),
		Name:        info.Name,
		Artists:     getArtistNames(info),
		Duration:    info.Duration,
		PreviewUrl:  info.PreviewURL,
		SuggestedBy: suggestingUser,
		Score:       score,
		TimeAdded:   time.Now(),
	}
}

func getArtistNames(songInfo *spotify.FullTrack) []string {
	names := make([]string, 0, len(songInfo.Artists))
	for _, artist := range songInfo.Artists {
		names = append(names, artist.Name)
	}
	return names
}
