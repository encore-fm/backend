package server

import (
	"github.com/antonbaumann/spotify-jukebox/handlers"
	"github.com/gorilla/mux"
)

func (s *Model) setupServerRoutes(r *mux.Router) {
	_ = r
}

func (s *Model) setupUserRoutes(r *mux.Router) {
	r.HandleFunc("/user/join", handlers.UserJoinHandler)
}
