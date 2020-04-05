package handlers

import (
	"errors"
	"net/http"

	log "github.com/sirupsen/logrus"
)

type FrontendError struct {
	Error       string `json:"error"`
	Description string `json:"description"`
}

var (
	// Vote errors
	ErrBadVoteAction = errors.New(`vote action must be in {"up", "down"}`)
	// Authentication Errors
	ErrWrongUserSecret = errors.New("user secret wrong")
	ErrUserNotAdmin    = errors.New("user not an admin")

	// Frontend errors
	SpotifyNotAuthenticated = FrontendError{
		Error:       "Spotify not authenticated",
		Description: "The Spotify authentication token has been generated for the requested user.",
	}

	RequestNotAuthorized = FrontendError{
		Error:       "Request not authorized",
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
	UserNotFoundError = FrontendError{
		Error:       "UserNotFoundError",
		Description: "No user with the specified ID exists.",
	}
	VoteOnSuggestedSongError = FrontendError{
		Error:       "VoteOnSuggestedSongError",
		Description: "The user requesting the vote has suggested this song.",
	}
	UserAlreadyVotedError = FrontendError{
		Error:       "UserAlreadyVotedError",
		Description: "The user requesting the vote has already voted for this song.",
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
