package handlers

import (
	"errors"
	"net/http"

	"github.com/antonbaumann/spotify-jukebox/config"
	log "github.com/sirupsen/logrus"
	"github.com/zmb3/spotify"
)

type SpotifyHandler struct {
	Authenticator spotify.Authenticator
	Client        spotify.Client
}

var ErrCouldNotGetToken = errors.New("couldn't get token")

// the user will eventually be redirected back to your redirect URL
func (h *SpotifyHandler) Redirect(w http.ResponseWriter, r *http.Request) {
	// use the same state string here that you used to generate the URL
	token, err := h.Authenticator.Token(config.Conf.Spotify.State, r)
	if err != nil {
		log.Errorf("handle spotify redirect: %v", ErrCouldNotGetToken)
		http.Error(w, ErrCouldNotGetToken.Error(), http.StatusNotFound)
		return
	}
	// create a client using the specified token
	h.Client = h.Authenticator.NewClient(token)
	// the client can now be used to make authenticated requests
	log.Info("spotify client can now be used to make authenticated requests")
	http.Redirect(w, r, config.Conf.Server.FrontendBaseUrl, http.StatusSeeOther)
}
