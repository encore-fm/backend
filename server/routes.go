package server

import (
	"net/http"

	"github.com/gorilla/mux"
)

func (s *Model) setupServerRoutes(r *mux.Router) {
	r.Handle(
		"/ping",
		http.HandlerFunc(s.Handler.Ping),
	)
}

func (s *Model) setupUserRoutes(r *mux.Router, auth authFunc) {
	r.Handle(
		"/users/join/{username}",
		http.HandlerFunc(s.Handler.Join),
	).Methods(http.MethodGet)

	r.Handle(
		"/users/{username}/list",
		auth(http.HandlerFunc(s.Handler.ListUsers)),
	).Methods(http.MethodGet)

	r.Handle(
		"/users/{username}/suggest/{song_id}",
		auth(http.HandlerFunc(s.Handler.SuggestSong)),
	).Methods(http.MethodGet)
}

func (s *Model) setupAdminRoutes(r *mux.Router) {
	r.Handle(
		"/admin/login",
		http.HandlerFunc(s.Handler.Login),
	).Methods(http.MethodPost)
}

func (s *Model) setupSpotifyRoutes(r *mux.Router) {
	r.Handle(
		"/callback",
		http.HandlerFunc(s.Handler.Redirect),
	)
}
