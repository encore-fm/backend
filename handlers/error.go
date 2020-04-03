package handlers

import (
	"errors"
	"fmt"
	"net/http"

	log "github.com/sirupsen/logrus"
)

type FrontendError struct {
	Code        string `json:"code"`
	Error       string `json:"error"`
	Description string `json:"description"`
}

var (
	// Spotify errors
	ErrSpotifyNotAuthenticated = errors.New("spotify not authenticated")
	// Vote errors
	ErrBadVoteAction = errors.New(`vote action must be in {"up", "down"}`)

	// Frontend errors
	SessionNotFoundError = FrontendError{
		Code:        "001",
		Error:       "Session not found",
		Description: "No session with the specified ID exists.",
	}
	SongNotFoundError = FrontendError{
		Code:        "002",
		Error:       "Song not found",
		Description: "No song with the specified ID exists.",
	}
	SessionConflictError = FrontendError{
		Code:        "003",
		Error:       "Session already exists",
		Description: "A session with the given ID already exists.",
	}
	UserConflictError = FrontendError{
		Code:        "004",
		Error:       "Username already exists",
		Description: "A user with the given username already exists.",
	}
	SongConflictError = FrontendError{
		Code:        "005",
		Error:       "Song already suggested",
		Description: "The song being requested has already been suggested in this session.",
	}
	BadVoteError = FrontendError{
		Code:        "006",
		Error:       "Bad vote request",
		Description: `Vote action specified in vote request has to be either "up" or "down"`,
	}
	UserNotFoundError = FrontendError{
		Code:        "007",
		Error:       "User not found",
		Description: "No user with the specified ID exists.",
	}
	VoteOnSuggestedSongError = FrontendError{
		Code:        "008",
		Error:       "User suggested song",
		Description: "The user requesting the vote has suggested this song.",
	}
	UserAlreadyVotedError = FrontendError{
		Code:        "009",
		Error:       "User already voted",
		Description: "The user requesting the vote has already voted for this song.",
	}
	InternalServerError = FrontendError{
		Code:        "010",
		Error:       "Internal server error",
		Description: "An unexpected server error has occurred.",
	}
)

func handleError(w http.ResponseWriter, status int, logLevel log.Level, msg string, err error, frontendError FrontendError) {
	switch logLevel {
	case log.WarnLevel:
		log.Warn(fmt.Sprintf(msg+": %v", err))
		break
	case log.ErrorLevel:
		log.Error(fmt.Sprintf(msg+": %v", err))
		break
	default:
		log.Info(fmt.Sprintf(msg+": %v", err))
		break
	}
	log.Errorf("%v: %v", err, status)
	jsonResponseWithStatus(w, status, frontendError)
}
