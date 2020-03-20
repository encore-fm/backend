package handlers

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/antonbaumann/spotify-jukebox/config"
	"github.com/antonbaumann/spotify-jukebox/db"
	"github.com/antonbaumann/spotify-jukebox/user"
	log "github.com/sirupsen/logrus"
)

type AdminHandler struct {
	UserCollection *db.UserCollection
}

var (
	ErrUserNotAdmin  = errors.New("user not an admin")
	ErrWrongPassword = errors.New("wrong password")
)

// Log in checks credentials {username, password} in request.Body
// if they match with configured admin credentials the admin-user
// struct will be returned
func (h *AdminHandler) Login(w http.ResponseWriter, r *http.Request) {
	var credentials config.AdminConfig
	if err := json.NewDecoder(r.Body).Decode(&credentials); err != nil {
		log.Errorf("admin login: decoding request body: %v", err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	if credentials.Username != config.Conf.Admin.Username {
		log.Errorf("admin login: %v", ErrUserNotAdmin)
		http.Error(w, ErrUserNotAdmin.Error(), http.StatusForbidden)
		return
	}
	if credentials.Password != config.Conf.Admin.Password {
		log.Errorf("admin login: %v", ErrWrongPassword)
		http.Error(w, ErrWrongPassword.Error(), http.StatusForbidden)
		return
	}

	u, err := h.UserCollection.GetUser(credentials.Username)
	if err != nil {
		log.Errorf("admin login: get user from db: %v", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if u != nil {
		log.Infof("admin login: [%v] successfully logged in", credentials.Username)
		jsonResponse(w, u)
		return
	}

	// if user does not exist in database -> create new user
	u, err = user.New(credentials.Username)
	if err != nil {
		log.Errorf("admin login: create new user: %v", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if err := h.UserCollection.AddUser(u); err != nil {
		log.Errorf("admin login: add user to db: %v", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	log.Infof("admin login: [%v] successfully logged in", credentials.Username)
	jsonResponse(w, u)
}
