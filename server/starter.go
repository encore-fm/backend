package server

import (
	"fmt"
	"net/http"
	"time"

	"github.com/antonbaumann/spotify-jukebox/config"
	"github.com/gorilla/handlers"
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
	s.setupUserRoutes(r, userAuth(s.UserCollection))
	s.setupAdminRoutes(r, adminAuth(s.UserCollection))
	s.setupEventRoutes(r)

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
	allowedOrigins := handlers.AllowedOrigins([]string{config.Conf.Server.FrontendBaseUrl})
	allowedHeaders := handlers.AllowedHeaders([]string{"X-Requested-With", "Content-Type", "Authorization", "Session"})
	allowedMethods := handlers.AllowedMethods([]string{"GET", "POST", "PUT", "DELETE", "HEAD", "OPTIONS"})
	err := http.ListenAndServe(addr, handlers.CORS(allowedOrigins, allowedHeaders, allowedMethods)(r))
	if err != nil {
		log.Errorf("server error: %v", err)
	}
}
