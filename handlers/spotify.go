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
	ErrUserNotPremium  = errors.New("user doesnt have a premium account")
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
		redirect(w, r, false, "")
		return
	}
	if usr == nil {
		log.Errorf("%v: %v", msg, ErrNoUserWithState)
		redirect(w, r, false, "")
		return
	}

	// error in authentication flow
	// occurs when user doesn't authorize app e.g. 'cancels'
	if authorizationErr != "" {
		log.Warnf("%v: %v", msg, authorizationErr)

		if usr.IsAdmin {
			h.deleteSessionAndUsers(ctx, usr.SessionID)

			// if user is admin dont redirect to callback popup
			http.Redirect(w, r, config.Conf.Server.FrontendBaseUrl, http.StatusSeeOther)
			return
		}

		redirect(w, r, true, "")
		return
	}

	// extract code from url
	code := values.Get("code")

	// use code to receive token
	token, err := h.spotifyAuthenticator.Exchange(code)
	if err != nil {
		log.Errorf("%v: %v", msg, err)
		redirect(w, r, false, "")
		return
	}

	// save token in user struct in db
	// this also sets the `SpotifyAuthorized` field to true
	if err := h.UserCollection.SetToken(ctx, usr.ID, token); err != nil {
		log.Errorf("%v: %v", msg, err)
		redirect(w, r, false, "")
		return
	}

	client := h.spotifyAuthenticator.NewClient(token)
	spotifyUser, err := client.CurrentUser()
	if err != nil {
		log.Errorf("%v: %v", msg, err)
		redirect(w, r, false, "")
		return
	}
	if spotifyUser.Product != "premium" {

		log.Errorf("%v: %v", msg, ErrUserNotPremium)
		redirect(w, r, false, "encore requires a valid spotify premium account.")
		return
	}

	// synchronize the user
	_, err = h.UserCollection.SetSynchronized(ctx, usr.ID, true)
	if err != nil {
		log.Errorf("%v: %v", msg, err)
	}
	h.eventBus.Publish(
		playerctrl.Synchronize,
		events.GroupID(usr.SessionID),
		playerctrl.SynchronizePayload{UserID: usr.ID},
	)

	redirectUrl := config.Conf.Server.FrontendBaseUrl
	if !usr.IsAdmin {
		redirectUrl = fmt.Sprintf("%v/callback/success", config.Conf.Server.FrontendBaseUrl)
	}

	log.Infof("%v: successfully received token for user [%v]", msg, usr.Username)
	http.Redirect(w, r, redirectUrl, http.StatusSeeOther)
}

func redirect(w http.ResponseWriter, r *http.Request, success bool, msg string) {
	successUrl := fmt.Sprintf("%v/callback/success", config.Conf.Server.FrontendBaseUrl)
	failureUrl := fmt.Sprintf("%v/callback/error", config.Conf.Server.FrontendBaseUrl)

	redirectUrl := successUrl
	if !success {
		redirectUrl = failureUrl
	}

	if msg != "" {
		redirectUrl = fmt.Sprintf("%v/%v", redirectUrl, msg)
	}
	http.Redirect(w, r, redirectUrl, http.StatusSeeOther)
}

func (h *handler) deleteSessionAndUsers(ctx context.Context, sessionID string) {
	msg := "[handler] redirect: delete session and users"
	err := h.UserCollection.DeleteUsersBySessionID(ctx, sessionID)
	if err != nil {
		log.Errorf("%v: %v", msg, err)
	} else {
		log.Infof("%v: successfully deleted all users in session [%v]", msg, sessionID)
	}
	err = h.SessionCollection.DeleteSession(ctx, sessionID)
	if err != nil {
		log.Errorf("%v: %v", msg, err)
	} else {
		log.Infof("%v: successfully deleted session [%v]", msg, sessionID)
	}
}
