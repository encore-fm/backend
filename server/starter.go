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

func (s *Model) Start() {
	start := time.Now()
	r := mux.NewRouter()
	s.setupServerRoutes(r)
	s.setupUserRoutes(r)
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
	corsObj := handlers.AllowedOrigins([]string{config.Conf.Server.FrontendBaseUrl})
	err := http.ListenAndServe(addr, handlers.CORS(corsObj)(r))
	if err != nil {
		log.Errorf("server error: %v", err)
	}
}
