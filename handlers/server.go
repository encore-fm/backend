package handlers

import (
	"net/http"

	log "github.com/sirupsen/logrus"
)


func (h *Handler) Ping(w http.ResponseWriter, r *http.Request) {
	log.Info("PING")
	w.WriteHeader(http.StatusOK)
}