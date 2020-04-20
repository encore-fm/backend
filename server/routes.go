package server

import (
	"net/http"

	"github.com/antonbaumann/spotify-jukebox/handlers"
	"github.com/gorilla/mux"
)

func (s *Model) setupServerRoutes(r *mux.Router) {
	r.Handle(
		"/ping",
		http.HandlerFunc(s.ServerHandler.Ping),
	)
}

func (s *Model) setupUserRoutes(r *mux.Router, auth handlers.AuthFunc) {
	r.Handle(
		"/users/{username}/join/{session_id}",
		http.HandlerFunc(s.UserHandler.Join),
	).Methods(http.MethodPost)

	r.Handle(
		"/users/{username}/leave",
		auth(http.HandlerFunc(s.UserHandler.Leave)),
	).Methods(http.MethodDelete)

	r.Handle(
		"/users/{username}/ping",
		auth(http.HandlerFunc(s.UserHandler.UserPing)),
	)

	r.Handle(
		"/users/{username}/list",
		auth(http.HandlerFunc(s.UserHandler.ListUsers)),
	).Methods(http.MethodGet)

	r.Handle(
		"/users/{username}/suggest/{song_id}",
		auth(http.HandlerFunc(s.UserHandler.SuggestSong)),
	).Methods(http.MethodPost)

	r.Handle(
		"/users/{username}/listSongs",
		auth(http.HandlerFunc(s.UserHandler.ListSongs)),
	).Methods(http.MethodGet)

	r.Handle(
		"/users/{username}/vote/{song_id}/{vote_action}",
		auth(http.HandlerFunc(s.UserHandler.Vote)),
	).Methods(http.MethodPost)

	r.Handle(
		"/users/{username}/clientToken",
		auth(http.HandlerFunc(s.UserHandler.ClientToken)),
	).Methods(http.MethodGet)

	r.Handle(
		"/users/{username}/authToken",
		auth(http.HandlerFunc(s.UserHandler.AuthToken)),
	).Methods(http.MethodGet)

	r.Handle(
		"/sessionInfo/{session_id}",
		http.HandlerFunc(s.UserHandler.SessionInfo),
	).Methods(http.MethodGet)
}

func (s *Model) setupAdminRoutes(r *mux.Router, auth handlers.AuthFunc) {
	r.Handle(
		"/admin/{username}/createSession",
		http.HandlerFunc(s.AdminHandler.CreateSession),
	).Methods(http.MethodPost)

	r.Handle(
		"/admin/{username}/deleteSession",
		auth(http.HandlerFunc(s.AdminHandler.DeleteSession)),
	).Methods(http.MethodDelete)

	r.Handle(
		"/users/{username}/removeSong/{song_id}",
		auth(http.HandlerFunc(s.AdminHandler.RemoveSong)),
	).Methods(http.MethodDelete)
}

func (s *Model) setupSpotifyRoutes(r *mux.Router) {
	r.Handle(
		"/callback",
		http.HandlerFunc(s.SpotifyHandler.Redirect),
	)
}

func (s *Model) setupEventRoutes(r *mux.Router) {
	r.Handle(
		"/events/{username}/{session_id}",
		http.HandlerFunc(s.SSEHandler.ServeHTTP),
	)
}

func (s *Model) setupPlayerRoutes(r *mux.Router, auth handlers.AuthFunc) {
	r.Handle(
		"/users/{username}/player/play",
		auth(http.HandlerFunc(s.PlayerHandler.Play)),
	).Methods(http.MethodPost)

	r.Handle(
		"/users/{username}/player/pause",
		auth(http.HandlerFunc(s.PlayerHandler.Pause)),
	).Methods(http.MethodPost)

	r.Handle(
		"/users/{username}/player/skip",
		auth(http.HandlerFunc(s.PlayerHandler.Skip)),
	).Methods(http.MethodPost)

	r.Handle(
		"/users/{username}/player/seek/{position_ms}",
		auth(http.HandlerFunc(s.PlayerHandler.Seek)),
	).Methods(http.MethodPost)

	r.Handle(
		"/users/{username}/player/state",
		auth(http.HandlerFunc(s.PlayerHandler.GetState)),
	).Methods(http.MethodGet)

	r.Handle(
		"/users/{username}/player/synchronize",
		auth(http.HandlerFunc(s.PlayerHandler.Synchronize)),
	).Methods(http.MethodPost)

	r.Handle(
		"/users/{username}/player/desynchronize",
		auth(http.HandlerFunc(s.PlayerHandler.Desynchronize)),
	).Methods(http.MethodPost)
}

func (s *Model) setupDebugRoutes(r *mux.Router) {
	r.Handle(
		"/debug/reset_player_controller/{session_id}",
		http.HandlerFunc(s.DebugHandler.ResetControllerState),
	)
}
