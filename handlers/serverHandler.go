package handlers

import (
	"net/http"

	log "github.com/sirupsen/logrus"
)

type ServerHandler struct {}

func (h *ServerHandler) Ping(w http.ResponseWriter, r *http.Request) {
	log.Info("PING")
	w.WriteHeader(http.StatusOK)
}
