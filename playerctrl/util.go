package playerctrl

import (
	"fmt"

	"github.com/zmb3/spotify"
)

func TrackURI(songID string) spotify.URI {
	return spotify.URI(fmt.Sprintf("spotify:track:%v", songID))
}
