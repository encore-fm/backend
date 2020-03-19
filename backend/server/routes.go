package server

import (
	"github.com/gorilla/mux"
)

func (s *Model) setupServerRoutes(r *mux.Router) {
	r.HandleFunc("/ping", s.ServerHandler.Ping)
}

func (s *Model) setupUserRoutes(r *mux.Router) {
	r.HandleFunc("/users/join/{username}", s.UserHandler.Join)
}
