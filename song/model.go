package song

import (
	"time"

	"github.com/antonbaumann/spotify-jukebox/user"
	"github.com/zmb3/spotify"
)

type Model struct {
	ID          string      `json:"id" bson:"id"`
	Name        string      `json:"name" bson:"name"`
	Artists     []string    `json:"artists" bson:"artists"`
	Duration    int         `json:"duration_ms" bson:"duration_ms"`
	CoverUrl    string      `json:"cover_url" bson:"cover_url"`
	AlbumName   string      `json:"album_name" bson:"album_name"`
	PreviewUrl  string      `json:"preview_url" bson:"preview_url"`
	SuggestedBy string      `json:"suggested_by" bson:"suggested_by"`
	Score       float64     `json:"score" bson:"score"`
	TimeAdded   time.Time   `json:"time_added" bson:"time_added"`
	Upvoters    user.Voters `json:"upvoters" bson:"upvoters"`
	Downvoters  user.Voters `json:"downvoters" bson:"downvoters"`
}

func New(
	suggestingUser string,
	score float64,
	info *spotify.FullTrack,
) *Model {
	albumUrl := ""
	if len(info.Album.Images) != 0 {
		albumUrl = info.Album.Images[0].URL
	}

	return &Model{
		ID:          string(info.ID),
		Name:        info.Name,
		Artists:     getArtistNames(info),
		Duration:    info.Duration,
		CoverUrl:    albumUrl,
		AlbumName:   info.Album.Name,
		PreviewUrl:  info.PreviewURL,
		SuggestedBy: suggestingUser,
		Score:       score,
		TimeAdded:   time.Now(),
		Upvoters:    user.NewVoters(),
		Downvoters:  user.NewVoters(),
	}
}

func getArtistNames(songInfo *spotify.FullTrack) []string {
	names := make([]string, 0, len(songInfo.Artists))
	for _, artist := range songInfo.Artists {
		names = append(names, artist.Name)
	}
	return names
}
