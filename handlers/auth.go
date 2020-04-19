package handlers

import (
	"context"
	"errors"
	"fmt"
	"net/http"

	"github.com/antonbaumann/spotify-jukebox/db"
	"github.com/antonbaumann/spotify-jukebox/user"
	"github.com/gorilla/mux"
	log "github.com/sirupsen/logrus"
)

type AuthFunc = func(http.Handler) http.Handler

func authenticate(userCollection db.UserCollection, checkAdmin bool) AuthFunc {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx := context.Background()

			userOrAdmin := "user"
			if checkAdmin {
				userOrAdmin = "admin"
			}
			msg := fmt.Sprintf("[auth] authenticate %v request", userOrAdmin)

			vars := mux.Vars(r)
			username := vars["username"]

			secret := r.Header.Get("Authorization")
			sessID := r.Header.Get("Session")
			userID := user.GenerateUserID(username, sessID)

			u, err := userCollection.GetUserByID(ctx, userID)
			// error while looking up user
			if errors.Is(err, db.ErrNoUserWithID) {
				handleError(w, http.StatusUnauthorized, log.WarnLevel, msg, err, RequestNotAuthorizedError)
				return
			}
			if err != nil {
				handleError(w, http.StatusInternalServerError, log.ErrorLevel, msg, err, InternalServerError)
				return
			}

			if secret != u.Secret {
				handleError(w, http.StatusUnauthorized, log.WarnLevel, msg, ErrWrongUserSecret, RequestNotAuthorizedError)
				return
			}

			if checkAdmin && !u.IsAdmin {
				handleError(w, http.StatusUnauthorized, log.WarnLevel, msg, ErrUserNotAdmin, ActionNotAllowedError)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

func UserAuth(userCollection db.UserCollection) AuthFunc {
	return authenticate(userCollection, false)
}

func AdminAuth(userCollection db.UserCollection) AuthFunc {
	return authenticate(userCollection, true)
}
