package handlers

import (
	"context"
	"errors"
	"fmt"
	"net/http"

	"github.com/antonbaumann/spotify-jukebox/events"
	"github.com/antonbaumann/spotify-jukebox/playerctrl"

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
	authorizationErr := values.Get("error")

	// find user linked to state
	usr, err := h.UserCollection.GetUserByState(ctx, actualState)
	if err != nil {
		log.Errorf("%v: %v", msg, err)
		http.Error(w, msg, http.StatusInternalServerError)
		return
	}
	if usr == nil {
		log.Errorf("%v: %v", msg, ErrNoUserWithState)
		http.Error(w, ErrNoUserWithState.Error(), http.StatusInternalServerError)
		return
	}

	// error in authentication flow
	// occurs when user doesn't authorize app e.g. 'cancels'
	if authorizationErr != "" {
		log.Warnf("%v: %v", msg, authorizationErr)

		if usr.IsAdmin {
			err = h.UserCollection.DeleteUsersBySessionID(ctx, usr.SessionID)
			if err != nil {
				log.Errorf("%v: %v", msg, err)
			} else {
				log.Infof("%v: successfully deleted all users in session [%v]", msg, usr.SessionID)
			}
			err = h.SessionCollection.DeleteSession(ctx, usr.SessionID)
			if err != nil {
				log.Errorf("%v: %v", msg, err)
			} else {
				log.Infof("%v: successfully deleted session [%v]", msg, usr.SessionID)
			}

			http.Redirect(w, r, config.Conf.Server.FrontendBaseUrl, http.StatusSeeOther)
			return
		}

		http.Redirect(
			w,
			r,
			fmt.Sprintf("%v/callback/error", config.Conf.Server.FrontendBaseUrl),
			http.StatusSeeOther,
		)
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
	if err := h.UserCollection.SetToken(ctx, usr.ID, token); err != nil {
		log.Errorf("%v: %v", msg, err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// synchronize the user
	_, err = h.UserCollection.SetSynchronized(ctx, usr.ID, true)
	if err != nil {
		log.Errorf("%v: %v", msg, err)
	}
	h.eventBus.Publish(
		playerctrl.SetSynchronizedEvent,
		events.GroupID(usr.SessionID),
		playerctrl.SetSynchronizedPayload{UserID: usr.ID, Synchronized: true},
	)

	redirectUrl := config.Conf.Server.FrontendBaseUrl
	if !usr.IsAdmin {
		redirectUrl = fmt.Sprintf("%v/callback/success", config.Conf.Server.FrontendBaseUrl)
	}

	log.Infof("%v: successfully received token for user [%v]", msg, usr.Username)
	http.Redirect(w, r, redirectUrl, http.StatusSeeOther)
}
