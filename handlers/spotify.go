package handlers

import (
	"errors"
	"net/http"

	"github.com/antonbaumann/spotify-jukebox/config"
	log "github.com/sirupsen/logrus"
)

var ErrCouldNotGetToken = errors.New("couldn't get token")

type SpotifyHandler interface {
	Redirect(w http.ResponseWriter, r *http.Request)
}

var _ SpotifyHandler = (*handler)(nil)

// the user will eventually be redirected back to your redirect URL
func (h *handler) Redirect(w http.ResponseWriter, r *http.Request) {
	// todo: extract state from url
	// todo: - find user linked to this state
	// todo: - extract code from url
	// todo: - get token
	// todo: - save token / client in user document
	//
	//values := r.URL.Query()
	//actualState := values.Get("state")
	//code := values.Get("code")
	//
	//token, err := h.spotifyAuthenticator.Exchange(code string)

	// use the same state string here that you used to generate the URL
	token, err := h.spotifyAuthenticator.Token(config.Conf.Spotify.State, r)
	if err != nil {
		log.Errorf("handle spotify redirect: %v", ErrCouldNotGetToken)
		http.Error(w, ErrCouldNotGetToken.Error(), http.StatusNotFound)
		return
	}
	// create a client using the specified token
	h.Spotify = h.spotifyAuthenticator.NewClient(token)
	h.spotifyActivated = true
	// the client can now be used to make authenticated requests
	log.Info("spotify client can now be used to make authenticated requests")
	http.Redirect(w, r, config.Conf.Server.FrontendBaseUrl, http.StatusSeeOther)
}
