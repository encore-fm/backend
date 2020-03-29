package handlers

import (
	"context"
	"errors"
	"net/http"

	"github.com/antonbaumann/spotify-jukebox/db"
	"github.com/antonbaumann/spotify-jukebox/session"
	"github.com/antonbaumann/spotify-jukebox/sse"
	"github.com/antonbaumann/spotify-jukebox/user"
	"github.com/gorilla/mux"
	log "github.com/sirupsen/logrus"
)

var (
	ErrWrongCredentials = errors.New("username and password do not match")
)

type AdminHandler interface {
	CreateSession(w http.ResponseWriter, r *http.Request)
	RemoveSong(w http.ResponseWriter, r *http.Request)
}

var _ AdminHandler = (*handler)(nil)

func (h *handler) CreateSession(w http.ResponseWriter, r *http.Request) {
	msg := "[handler] create session"
	ctx := context.Background()

	vars := mux.Vars(r)
	username := vars["username"]

	// create new session (contains random session id)
	sess, err := session.New()
	if err != nil {
		log.Errorf("%v: creating new session: %v", msg, err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// save session in db
	err = h.SessionCollection.AddSession(ctx, sess)
	if err != nil {
		if errors.Is(err, db.ErrSessionAlreadyExisting) {
			log.Errorf("%v: saving session: %v", msg, err)
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		log.Errorf("%v: saving session: %v", msg, err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// create admin user. contains
	// - user secret
	// - state for spotify authentication
	admin, err := user.NewAdmin(username, sess.ID)
	if err != nil {
		log.Errorf("%v: create admin user: %v", msg, err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// save admin user in db
	if err := h.UserCollection.AddUser(ctx, admin); err != nil {
		log.Errorf("%v: save admin user: %v", msg, err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// create authentication url containing auth state
	// auth state will later be used to link spotify callback to user
	authUrl := h.spotifyAuthenticator.AuthURLWithDialog(admin.AuthState)

	response := &struct {
		UserInfo *user.Model `json:"user_info"`
		AuthUrl  string      `json:"auth_url"`
	}{
		UserInfo: admin,
		AuthUrl:  authUrl,
	}

	log.Infof("%v: [%v] successfully created session with id [%v]", msg, username, sess.ID)
	jsonResponse(w, response)
}

func (h *handler) RemoveSong(w http.ResponseWriter, r *http.Request) {
	ctx := context.Background()
	msg := "remove song"
	vars := mux.Vars(r)
	songID := vars["song_id"]

	if err := h.SongCollection.RemoveSong(ctx, songID); err != nil {
		log.Errorf("%v: %v", msg, err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	songList, err := h.SongCollection.ListSongs(ctx)
	if err != nil {
		log.Errorf("%v: %v", msg, err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// send new playlist to broker
	event := sse.Event{
		Event: sse.PlaylistChange,
		Data:  songList,
	}
	h.Broker.Notifier <- event

	log.Infof("admin removed song [%v]", songID)
	jsonResponse(w, songList)
}
