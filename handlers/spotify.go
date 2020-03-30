package handlers

import (
	"context"
	"errors"
	"net/http"

	"github.com/antonbaumann/spotify-jukebox/config"
	log "github.com/sirupsen/logrus"
)

var (
	ErrNoUserWithState = errors.New("no user with this state in db")
)

type SpotifyHandler interface {
	Redirect(w http.ResponseWriter, r *http.Request)
}

var _ SpotifyHandler = (*handler)(nil)

// the user will eventually be redirected back to your redirect URL
func (h *handler) Redirect(w http.ResponseWriter, r *http.Request) {
	msg := "[handler] redirect"
	ctx := context.Background()

	values := r.URL.Query()

	// extract actual state from url
	actualState := values.Get("state")

	// find user linked to state
	user, err := h.UserCollection.GetUserByState(ctx, actualState)
	if err != nil {
		log.Errorf("%v: %v", msg, err)
		http.Error(w, msg, http.StatusInternalServerError)
		return
	}
	if user == nil {
		log.Errorf("%v: %v", msg, ErrNoUserWithState)
		http.Error(w, ErrNoUserWithState.Error(), http.StatusInternalServerError)
		return
	}

	// extract code from url
	code := values.Get("code")

	// use code to receive token
	token, err := h.spotifyAuthenticator.Exchange(code)
	if err != nil {
		log.Errorf("%v: %v", msg, err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// save token in user struct in db
	// this also sets the `SpotifyAuthorized` field to true
	if err := h.UserCollection.SetToken(ctx, user.ID, token); err != nil {
		log.Errorf("%v: %v", msg, err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	log.Infof("%v: successfully received token for user [%v]", msg, user.Username)
	http.Redirect(w, r, config.Conf.Server.FrontendBaseUrl, http.StatusSeeOther)
}
