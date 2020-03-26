package handlers

import (
	"net/http"

	log "github.com/sirupsen/logrus"
)

type ServerHandler interface {
	Ping(w http.ResponseWriter, r *http.Request)
}

var _ ServerHandler = (*handler)(nil)

func (h *handler) Ping(w http.ResponseWriter, r *http.Request) {
	log.Info("PING")
	w.WriteHeader(http.StatusOK)
}