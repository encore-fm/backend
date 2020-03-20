package server

import (
	"net/http"

	"github.com/gorilla/mux"
)

func (s *Model) setupServerRoutes(r *mux.Router) {
	r.HandleFunc("/ping", s.ServerHandler.Ping)
}

func (s *Model) setupUserRoutes(r *mux.Router) {
	r.HandleFunc("/users/join/{username}", s.UserHandler.Join).Methods(http.MethodGet)
	r.HandleFunc("/users/{username}/list", s.UserHandler.ListUsers).Methods(http.MethodGet)
}

func (s *Model) setupAdminRoutes(r *mux.Router) {
	r.HandleFunc("/admin/login", s.AdminHandler.Login).Methods(http.MethodPost)
}
