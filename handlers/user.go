package handlers

import (
	"context"
	"errors"
	"net/http"

	"github.com/antonbaumann/spotify-jukebox/playerctrl"

	"github.com/antonbaumann/spotify-jukebox/db"
	"github.com/antonbaumann/spotify-jukebox/events"
	"github.com/antonbaumann/spotify-jukebox/song"
	"github.com/antonbaumann/spotify-jukebox/sse"
	"github.com/antonbaumann/spotify-jukebox/user"
	"github.com/gorilla/mux"
	log "github.com/sirupsen/logrus"
	"github.com/zmb3/spotify"
)

type UserHandler interface {
	Join(w http.ResponseWriter, r *http.Request)
	Leave(w http.ResponseWriter, r *http.Request)
	UserInfo(w http.ResponseWriter, r *http.Request)
	UserPing(w http.ResponseWriter, r *http.Request)
	ListUsers(w http.ResponseWriter, r *http.Request)
	SuggestSong(w http.ResponseWriter, r *http.Request)
	ListSongs(w http.ResponseWriter, r *http.Request)
	Vote(w http.ResponseWriter, r *http.Request)
	ClientToken(w http.ResponseWriter, r *http.Request)
	AuthToken(w http.ResponseWriter, r *http.Request)
	SessionInfo(w http.ResponseWriter, r *http.Request)
}

var _ UserHandler = (*handler)(nil)

// Join adds user to session
// - check if session with this id exists
// - create new user and save in db
// - create auth url
func (h *handler) Join(w http.ResponseWriter, r *http.Request) {
	msg := "[handler] join"
	ctx := context.Background()

	vars := mux.Vars(r)
	username := vars["username"]
	sessionID := vars["session_id"]

	// check if session with given id exists
	sess, err := h.SessionCollection.GetSessionByID(ctx, sessionID)
	if err != nil {
		if errors.Is(err, db.ErrNoSessionWithID) {
			handleError(w, http.StatusBadRequest, log.ErrorLevel, msg, err, SessionNotFoundError)
		} else {
			handleError(w, http.StatusInternalServerError, log.ErrorLevel, msg, err, InternalServerError)
		}
		return
	}

	// update session time stamp
	h.SessionCollection.SetLastUpdated(ctx, sessionID)

	newUser, err := user.New(username, sessionID)
	if err != nil {
		if errors.Is(err, user.ErrUsernameTooShort) {
			handleError(w, http.StatusBadRequest, log.DebugLevel, msg, err, UsernameTooShortError)
			return
		}
		if errors.Is(err, user.ErrUsernameTooLong) {
			handleError(w, http.StatusBadRequest, log.DebugLevel, msg, err, UsernameTooLongError)
			return
		}
		if errors.Is(err, user.ErrUsernameInvalidCharacter) {
			handleError(w, http.StatusBadRequest, log.DebugLevel, msg, err, UsernameInvalidCharacterError)
			return
		}

		handleError(w, http.StatusInternalServerError, log.ErrorLevel, msg, err, InternalServerError)
		return
	}

	// save user in db
	if err := h.UserCollection.AddUser(ctx, newUser); err != nil {
		if errors.Is(err, db.ErrUsernameTaken) {
			handleError(w, http.StatusConflict, log.ErrorLevel, msg, err, UserConflictError)
		} else {
			handleError(w, http.StatusInternalServerError, log.ErrorLevel, msg, err, InternalServerError)
		}
		return
	}

	// create authentication url containing auth state
	// auth state will later be used to link spotify callback to user
	authUrl := h.spotifyAuthenticator.AuthURLWithDialog(newUser.AuthState)

	response := &struct {
		UserInfo *user.Model `json:"user_info"`
		AuthUrl  string      `json:"auth_url"`
	}{
		UserInfo: newUser,
		AuthUrl:  authUrl,
	}

	log.Infof("%v: [%v] successfully joined session with id [%v]", msg, username, sess.ID)
	jsonResponse(w, response)
}

func (h *handler) Leave(w http.ResponseWriter, r *http.Request) {
	msg := "[handler] leave"
	ctx := context.Background()

	vars := mux.Vars(r)
	username := vars["username"]
	sessionID := r.Header.Get("Session")
	userID := user.GenerateUserID(username, sessionID)

	usr, err := h.UserCollection.GetUserByID(ctx, userID)
	if err != nil {
		handleError(w, http.StatusInternalServerError, log.ErrorLevel, msg, err, InternalServerError)
		return
	}
	// admin should not be allowed the leave his own session (see delete session endpoint)
	if usr.IsAdmin {
		handleError(w, http.StatusBadRequest, log.WarnLevel, msg, ErrUserIsAdmin, ActionNotAllowedError)
		return
	}

	// pause the user's spotify client
	if usr.SpotifySynchronized {
		clients, err := h.UserCollection.GetSpotifyClients(ctx, sessionID)
		if err != nil {
			log.Errorf("%v: %v", msg, err)
		}
		// find user's client
		for _, client := range clients {
			if client.ID == userID {
				spotifyClient := h.spotifyAuthenticator.NewClient(client.AuthToken)
				err = spotifyClient.Pause()
				if err != nil {
					log.Errorf("%v: %v", msg, err)
				}
				break
			}
		}
	}

	err = h.UserCollection.DeleteUser(ctx, userID)
	if err != nil {
		handleError(w, http.StatusInternalServerError, log.ErrorLevel, msg, err, InternalServerError)
	}
}

func (h *handler) UserInfo(w http.ResponseWriter, r *http.Request) {
	msg := "[handler] user info"
	ctx := context.Background()

	vars := mux.Vars(r)
	username := vars["username"]
	sessionID := r.Header.Get("Session")

	userID := user.GenerateUserID(username, sessionID)
	usr, err := h.UserCollection.GetUserByID(ctx, userID)
	if err != nil {
		handleError(w, http.StatusInternalServerError, log.ErrorLevel, msg, err, InternalServerError)
		return
	}

	jsonResponse(w, usr)
}

func (h *handler) UserPing(w http.ResponseWriter, r *http.Request) {
	msg := "[handler] ping"
	vars := mux.Vars(r)
	username := vars["username"]

	response := &struct {
		Message string `json:"message"`
	}{
		Message: "pong",
	}
	jsonResponse(w, response)
	log.Infof("%v: %v", msg, username)
}

// ListUsers lists all users in the session
func (h *handler) ListUsers(w http.ResponseWriter, r *http.Request) {
	msg := "[handler] list users"
	ctx := context.Background()

	vars := mux.Vars(r)
	username := vars["username"]
	sessionID := r.Header.Get("Session")

	userList, err := h.UserCollection.ListUsers(ctx, sessionID)
	if err != nil {
		handleError(w, http.StatusInternalServerError, log.ErrorLevel, msg, err, InternalServerError)
		return
	}

	log.Infof("user [%v] requested user list", username)
	jsonResponse(w, userList)
}

// ListUsers add's a new song to the song_list
func (h *handler) SuggestSong(w http.ResponseWriter, r *http.Request) {
	msg := "[handler] suggest song"
	ctx := context.Background()

	vars := mux.Vars(r)
	username := vars["username"]
	songID := vars["song_id"]
	sessionID := r.Header.Get("Session")

	// update session time stamp
	h.SessionCollection.SetLastUpdated(ctx, sessionID)

	fullTrack, err := h.Spotify.Client.GetTrack(spotify.ID(songID))
	if err != nil {
		// todo: should mostly be UserError -> better checks
		handleError(w, http.StatusInternalServerError, log.ErrorLevel, msg, err, InternalServerError)
		return
	}

	// if user suggest's song he automatically votes up
	songInfo := song.New(username, 1, fullTrack)
	songInfo.Upvoters = append(songInfo.Upvoters, username)

	if err := h.SongCollection.AddSong(ctx, sessionID, songInfo); err != nil {
		if errors.Is(err, db.ErrSongAlreadyInSession) {
			handleError(w, http.StatusConflict, log.WarnLevel, msg, err, SongConflictError)
		} else {
			handleError(w, http.StatusInternalServerError, log.ErrorLevel, msg, err, InternalServerError)
		}
		return
	}

	log.Infof("%v: by [%v] songID [%v]", msg, username, songID)
	jsonResponse(w, songInfo)

	// fetch songList and send event
	songList, err := h.SongCollection.ListSongs(ctx, sessionID)
	if err != nil {
		log.Errorf("%v: event: %v", msg, err)
	}

	h.eventBus.Publish(sse.PlaylistChange, events.GroupID(sessionID), songList)
	// notify the player controller of a new song being suggested
	h.eventBus.Publish(playerctrl.SongAdded, events.GroupID(sessionID), nil)
}

// ListSongs returns all songs in one session
func (h *handler) ListSongs(w http.ResponseWriter, r *http.Request) {
	msg := "[handler] list songs"
	ctx := context.Background()

	vars := mux.Vars(r)
	username := vars["username"]

	sessionID := r.Header.Get("Session")

	// update session time stamp
	h.SessionCollection.SetLastUpdated(ctx, sessionID)

	songList, err := h.SongCollection.ListSongs(ctx, sessionID)
	if err != nil {
		handleError(w, http.StatusInternalServerError, log.ErrorLevel, msg, err, InternalServerError)
		return
	}

	// dont return nil if SongList is empty
	if songList == nil {
		songList = make([]*song.Model, 0)
	}

	log.Infof("%v: user [%v]", msg, username)
	jsonResponse(w, songList)
}

// Vote handles user votes on songs
func (h *handler) Vote(w http.ResponseWriter, r *http.Request) {
	msg := "[handler] vote"
	ctx := context.Background()

	vars := mux.Vars(r)
	username := vars["username"]
	songID := vars["song_id"]
	voteAction := vars["vote_action"]

	// get session id from headers
	sessionID := r.Header.Get("Session")

	// update session time stamp
	h.SessionCollection.SetLastUpdated(ctx, sessionID)

	if voteAction != "up" && voteAction != "down" {
		handleError(w, http.StatusBadRequest, log.ErrorLevel, msg, ErrBadVoteAction, BadVoteError)
		return
	}

	// the scoreChange of the song score has to be applied to the user score
	var scoreChange int
	var err error

	if voteAction == "up" {
		scoreChange, err = h.SongCollection.VoteUp(ctx, sessionID, songID, username)
		if err != nil {
			handleError(w, http.StatusInternalServerError, log.ErrorLevel, msg, err, InternalServerError)
			return
		}
	} else {
		scoreChange, err = h.SongCollection.VoteDown(ctx, sessionID, songID, username)
		if err != nil {
			handleError(w, http.StatusInternalServerError, log.ErrorLevel, msg, err, InternalServerError)
			return
		}
	}

	songInfo, err := h.SongCollection.GetSongByID(ctx, sessionID, songID)
	if err != nil {
		handleError(w, http.StatusInternalServerError, log.ErrorLevel, msg, err, InternalServerError)
		return
	}

	// if user updates his vote on own song -> dont update his score
	if songInfo.SuggestedBy != username {
		// update score of user that suggested song
		if err := h.UserCollection.IncrementScore(
			ctx,
			user.GenerateUserID(songInfo.SuggestedBy, sessionID),
			scoreChange,
		); err != nil {
			handleError(w, http.StatusInternalServerError, log.ErrorLevel, msg, err, InternalServerError)
			return
		}
	}

	// return updated song list
	songList, err := h.SongCollection.ListSongs(ctx, sessionID)
	if err != nil {
		handleError(w, http.StatusInternalServerError, log.ErrorLevel, msg, err, InternalServerError)
		return
	}

	log.Infof("user [%v] %vvoted song [%v]", username, voteAction, songID)
	jsonResponse(w, songList)

	h.eventBus.Publish(sse.PlaylistChange, events.GroupID(sessionID), songList)
}

// returns client token
func (h *handler) ClientToken(w http.ResponseWriter, r *http.Request) {
	msg := "[handler] get client token"
	ctx := context.Background()
	vars := mux.Vars(r)
	username := vars["username"]
	sessionID := r.Header.Get("Session")

	// update session time stamp
	h.SessionCollection.SetLastUpdated(ctx, sessionID)

	token, err := h.Spotify.GetClientToken()
	if err != nil {
		handleError(w, http.StatusInternalServerError, log.ErrorLevel, msg, err, InternalServerError)
		return
	}

	// don't return refresh token to frontend
	token.RefreshToken = ""
	jsonResponse(w, token)
	log.Infof("%v: user=[%v], session=[%v]", msg, username, sessionID)
}

func (h *handler) AuthToken(w http.ResponseWriter, r *http.Request) {
	msg := "[handler] get auth token"
	ctx := context.Background()

	vars := mux.Vars(r)
	username := vars["username"]
	sessionID := r.Header.Get("Session")

	// update session time stamp
	h.SessionCollection.SetLastUpdated(ctx, sessionID)

	userID := user.GenerateUserID(username, sessionID)
	usr, err := h.UserCollection.GetUserByID(ctx, userID)
	if err != nil {
		if errors.Is(err, db.ErrNoUserWithID) {
			handleError(w, http.StatusUnauthorized, log.WarnLevel, msg, err, RequestNotAuthorizedError)
		} else {
			handleError(w, http.StatusInternalServerError, log.ErrorLevel, msg, err, InternalServerError)
		}
		return
	}

	if !usr.SpotifyAuthorized {
		handleError(w, http.StatusUnauthorized, log.WarnLevel, msg, ErrSpotifyNotAuthenticated, SpotifyNotAuthenticatedError)
		return
	}

	token := usr.AuthToken
	// don't send refresh token
	token.RefreshToken = ""

	jsonResponse(w, token)
	log.Infof("%v: user=[%v], session=[%v]", msg, username, sessionID)
}

func (h *handler) SessionInfo(w http.ResponseWriter, r *http.Request) {
	msg := "[handler] get session info"
	ctx := context.Background()

	vars := mux.Vars(r)
	sessionID := vars["session_id"]

	// get session's admin
	admin, err := h.UserCollection.GetAdminBySessionID(ctx, sessionID)
	if err != nil {
		if errors.Is(err, db.ErrNoSessionWithID) {
			handleError(w, http.StatusBadRequest, log.WarnLevel, msg, err, SessionNotFoundError)
		} else {
			handleError(w, http.StatusInternalServerError, log.ErrorLevel, msg, err, InternalServerError)
		}
		return
	}

	// get session's player for current song
	player, err := h.PlayerCollection.GetPlayer(ctx, sessionID)
	if err != nil {
		if errors.Is(err, db.ErrNoSessionWithID) {
			handleError(w, http.StatusBadRequest, log.WarnLevel, msg, err, SessionNotFoundError)
		} else {
			handleError(w, http.StatusInternalServerError, log.ErrorLevel, msg, err, InternalServerError)
		}
		return
	}
	var currentSong *song.Model
	if player != nil {
		currentSong = player.CurrentSong
	}

	response := &struct {
		AdminName   string      `json:"admin_name"`
		CurrentSong *song.Model `json:"current_song"`
	}{
		AdminName:   admin.Username,
		CurrentSong: currentSong,
	}

	jsonResponse(w, response)
	log.Infof("%v: session=[%v]", msg, sessionID)
}
