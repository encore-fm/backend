package server

import (
	"net/http"

	"github.com/gorilla/mux"
)

func (s *Model) setupServerRoutes(r *mux.Router) {
	r.HandleFunc("/ping", s.Handler.Ping)
}

func (s *Model) setupUserRoutes(r *mux.Router) {
	r.HandleFunc("/users/join/{username}", s.Handler.Join).Methods(http.MethodGet)
	r.HandleFunc("/users/{username}/list", s.Handler.ListUsers).Methods(http.MethodGet)

	r.HandleFunc("/users/test", s.Handler.Test).Methods(http.MethodGet)
}

func (s *Model) setupAdminRoutes(r *mux.Router) {
	r.HandleFunc("/admin/login", s.Handler.Login).Methods(http.MethodPost)
}

func (s *Model) setupSpotifyRoutes(r *mux.Router) {
	r.HandleFunc("/callback", s.Handler.Redirect)
}
