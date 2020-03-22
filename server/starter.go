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
	s.setupUserRoutes(r, userAuth(s.Handler.UserCollection))
	s.setupAdminRoutes(r)

	http.Handle("/", r)

	log.Infof(
		"server started at port %v, took %v",
		s.Port,
		time.Since(start),
	)

	s.listenAndServe(r)
}

func (s *Model) listenAndServe(r *mux.Router) {
	addr := fmt.Sprintf(":%v", s.Port)
	allowedOrigins := handlers.AllowedOrigins([]string{config.Conf.Server.FrontendBaseUrl})
	allowedHeaders := handlers.AllowedHeaders([]string{"X-Requested-With", "Content-Type", "Authorization"})
	allowedMethods := handlers.AllowedMethods([]string{"GET", "POST", "PUT", "HEAD", "OPTIONS"})
	err := http.ListenAndServe(addr, handlers.CORS(allowedOrigins, allowedHeaders, allowedMethods)(r))
	if err != nil {
		log.Errorf("server error: %v", err)
	}
}
