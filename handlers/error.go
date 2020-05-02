package handlers

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/antonbaumann/spotify-jukebox/user"
	log "github.com/sirupsen/logrus"
)

type FrontendError struct {
	Error       string `json:"error"`
	Description string `json:"description"`
}

var (
	// Spotify errors
	ErrSpotifyNotAuthenticated = errors.New("spotify not authenticated")
	// Vote errors
	ErrBadVoteAction = errors.New(`vote action must be in {"up", "down"}`)
	// Authentication Errors
	ErrWrongUserSecret = errors.New("user secret wrong")
	ErrUserNotAdmin    = errors.New("user not an admin")
	// Actions that cannot be performed by the admin e.g. leaving session
	ErrUserIsAdmin = errors.New("the action cannot be performed by an admin")
	// specifies that a sync mode did not match expected form
	ErrBadSyncMode = errors.New(`sync mode must be in {"FORCE_SYNC", "FORCE_DESYNC", "AUTO"}`)

	// Frontend errors
	UsernameTooShortError = FrontendError{
		Error:       "UsernameTooShortError",
		Description: fmt.Sprintf("username should have at least %v characters.", user.MinLen),
	}
	UsernameTooLongError = FrontendError{
		Error:       "UsernameTooLongError",
		Description: fmt.Sprintf("username should not have more than %v characters.", user.MaxLen),
	}
	UsernameInvalidCharacterError = FrontendError{
		Error:       "UsernameInvalidCharacterError",
		Description: "username contains invalid character",
	}
	RequestUrlMalformedError = FrontendError{
		Error:       "RequestUrlMalformedError",
		Description: "Request url does not match expected model.",
	}
	RequestBodyMalformedError = FrontendError{
		Error:       "RequestBodyMalformedError",
		Description: "Request body does not match expected model.",
	}
	SpotifyNotAuthenticatedError = FrontendError{
		Error:       "SpotifyNotAuthenticatedError",
		Description: "No Spotify authentication token has been generated for the requested user.",
	}
	RequestNotAuthorizedError = FrontendError{
		Error:       "RequestNotAuthorizedError",
		Description: "Combination of username, sessionID and user secret is wrong",
	}
	SessionNotFoundError = FrontendError{
		Error:       "SessionNotFoundError",
		Description: "No session with the specified ID exists.",
	}
	SongNotFoundError = FrontendError{
		Error:       "SongNotFoundError",
		Description: "No song with the specified ID exists.",
	}
	SessionConflictError = FrontendError{
		Error:       "SessionConflictError",
		Description: "A session with the given ID already exists.",
	}
	UserConflictError = FrontendError{
		Error:       "UserConflictError",
		Description: "A user with the given username already exists.",
	}
	SongConflictError = FrontendError{
		Error:       "SongConflictError",
		Description: "The song being requested has already been suggested in this session.",
	}
	BadVoteError = FrontendError{
		Error:       "BadVoteError",
		Description: `Vote action specified in vote request has to be either "up" or "down"`,
	}
	BadSyncModeError = FrontendError{
		Error:       "BadSyncModeError",
		Description: `sync mode must be in {"FORCE_SYNC", "FORCE_DESYNC", "AUTO"}`,
	}
	UserNotFoundError = FrontendError{
		Error:       "UserNotFoundError",
		Description: "No user with the specified ID exists.",
	}
	ActionNotAllowedError = FrontendError{
		Error:       "ActionNotAllowedError",
		Description: "User does not have sufficient permissions to perform this action.",
	}
	InternalServerError = FrontendError{
		Error:       "InternalServerError",
		Description: "An unexpected server error has occurred.",
	}
)

func handleError(w http.ResponseWriter, status int, logLevel log.Level, msg string, err error, frontendError FrontendError) {
	switch logLevel {
	case log.WarnLevel:
		log.Warnf("%v: %v", msg, err)
		break
	case log.ErrorLevel:
		log.Errorf("%v: %v", msg, err)
		break
	default:
		log.Infof("%v: %v", msg, err)
		break
	}
	jsonResponseWithStatus(w, status, frontendError)
}
