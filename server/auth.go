package server

import (
	"context"
	"errors"
	"net/http"

	"github.com/antonbaumann/spotify-jukebox/db"
	"github.com/antonbaumann/spotify-jukebox/user"
	"github.com/gorilla/mux"
	log "github.com/sirupsen/logrus"
)

var (
	ErrUserNotFound    = errors.New("user not found")
	ErrWrongUserSecret = errors.New("wrong user secret")
	ErrUserNotAdmin    = errors.New("user not an admin")
)

type authFunc = func(http.Handler) http.Handler

func authenticate(userCollection db.UserCollection, checkAdmin bool) authFunc {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx := context.Background()
			msg := "authenticate admin request: %v"
			vars := mux.Vars(r)
			username := vars["username"]

			secret := r.Header.Get("Authorization")
			sessID := r.Header.Get("Session")
			userID := user.GenerateUserID(username, sessID)

			u, err := userCollection.GetUserByID(ctx, userID)
			// error while looking up user
			if err != nil {
				log.Errorf(msg, err)
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			// combination of username and session does not exist
			if u == nil {
				log.Warnf(msg, ErrUserNotFound)
				http.Error(w, ErrUserNotFound.Error(), http.StatusUnauthorized)
				return
			}

			if secret != u.Secret {
				log.Warnf(msg, ErrWrongUserSecret)
				http.Error(w, ErrWrongUserSecret.Error(), http.StatusUnauthorized)
				return
			}

			if checkAdmin && !u.IsAdmin {
				log.Warnf(msg, ErrUserNotAdmin)
				http.Error(w, ErrUserNotAdmin.Error(), http.StatusUnauthorized)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

func userAuth(userCollection db.UserCollection) authFunc {
	return authenticate(userCollection, false)
}

func adminAuth(userCollection db.UserCollection) authFunc {
	return authenticate(userCollection, true)
}
