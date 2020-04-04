package server

import (
	"fmt"
	"net/http"
	"time"

	"github.com/antonbaumann/spotify-jukebox/config"
	"github.com/antonbaumann/spotify-jukebox/handlers"
	muxh "github.com/gorilla/handlers"
	"github.com/gorilla/mux"
	log "github.com/sirupsen/logrus"
)

// Start sets up all routes and starts the server
func (s *Model) Start() {
	start := time.Now()
	r := mux.NewRouter()

	// setup routes
	s.setupServerRoutes(r)
	s.setupSpotifyRoutes(r)
	s.setupUserRoutes(r, handlers.UserAuth(s.UserCollection))
	s.setupAdminRoutes(r, handlers.AdminAuth(s.UserCollection))
	s.setupEventRoutes(r, handlers.UserAuth(s.UserCollection))

	http.Handle("/", r)

	log.Infof(
		"[startup] server started at port %v, took %v",
		s.Port,
		time.Since(start),
	)

	s.listenAndServe(r)
}

func (s *Model) listenAndServe(r *mux.Router) {
	addr := fmt.Sprintf(":%v", s.Port)
	allowedOrigins := muxh.AllowedOrigins([]string{config.Conf.Server.FrontendBaseUrl})
	allowedHeaders := muxh.AllowedHeaders([]string{"X-Requested-With", "Content-Type", "Authorization, Session"})
	allowedMethods := muxh.AllowedMethods([]string{"GET", "POST", "PUT", "DELETE", "HEAD", "OPTIONS"})
	err := http.ListenAndServe(addr, muxh.CORS(allowedOrigins, allowedHeaders, allowedMethods)(r))
	if err != nil {
		log.Errorf("server error: %v", err)
	}
}
