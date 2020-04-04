package handlers

import (
	"encoding/json"
	"net/http"

	log "github.com/sirupsen/logrus"
)

func jsonResponse(w http.ResponseWriter, v interface{}) {
	jsonResponseWithStatus(w, http.StatusOK, v)
}

func jsonResponseWithStatus(w http.ResponseWriter, statusCode int, v interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	err := json.NewEncoder(w).Encode(v)
	if err != nil {
		log.Errorf("write response body: %v", err)
	}
}
