package handlers

import (
	"net/http"

	"github.com/davecgh/go-spew/spew"
	"github.com/gorilla/mux"
	log "github.com/sirupsen/logrus"
)

func UserJoinHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	spew.Dump(vars)
	w.WriteHeader(http.StatusOK)
	log.Println("/user/join")
}
