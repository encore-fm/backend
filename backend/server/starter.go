package server

import (
	"fmt"
	"net/http"
	"strconv"
	"time"

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
		"server started at port %s, took %v",
		strconv.Itoa(s.Port),
		time.Since(start),
	)
	listenAndServe(s)
}

func listenAndServe(s *Model) {
	addr := fmt.Sprintf(":%v", s.Port)
	err := http.ListenAndServe(addr, nil)
	if err != nil {
		log.Errorf("server error: %v", err)
	}
}