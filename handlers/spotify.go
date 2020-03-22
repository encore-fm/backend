package handlers

import (
	"errors"
	"net/http"

	"github.com/antonbaumann/spotify-jukebox/config"
	log "github.com/sirupsen/logrus"
)


var ErrCouldNotGetToken = errors.New("couldn't get token")

// the user will eventually be redirected back to your redirect URL
func (h *Handler) Redirect(w http.ResponseWriter, r *http.Request) {
	// use the same state string here that you used to generate the URL
	token, err := h.spotifyAuthenticator.Token(config.Conf.Spotify.State, r)
	if err != nil {
		log.Errorf("handle spotify redirect: %v", ErrCouldNotGetToken)
		http.Error(w, ErrCouldNotGetToken.Error(), http.StatusNotFound)
		return
	}
	// create a client using the specified token
	h.Spotify = h.spotifyAuthenticator.NewClient(token)
	h.spotifyIsAuthenticated = true
	// the client can now be used to make authenticated requests
	log.Info("spotify client can now be used to make authenticated requests")
	http.Redirect(w, r, config.Conf.Server.FrontendBaseUrl, http.StatusSeeOther)
}
