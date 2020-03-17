package handlers

import (
	"net/http"

	log "github.com/sirupsen/logrus"
)

func UserJoinHandler(w http.ResponseWriter, r *http.Request) {
	//vars := mux.Vars(r)
	w.WriteHeader(http.StatusOK)
	log.Println("/user/join")
}
