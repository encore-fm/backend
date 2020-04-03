package handlers

import (
	"errors"
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
		Error:       "Session not found",
		Description: "No session with the specified ID exists.",
	}
	SongNotFoundError = FrontendError{
		Error:       "Song not found",
		Description: "No song with the specified ID exists.",
	}
	SessionConflictError = FrontendError{
		Error:       "Session already exists",
		Description: "A session with the given ID already exists.",
	}
	UserConflictError = FrontendError{
		Error:       "Username already exists",
		Description: "A user with the given username already exists.",
	}
	SongConflictError = FrontendError{
		Error:       "Song already suggested",
		Description: "The song being requested has already been suggested in this session.",
	}
	BadVoteError = FrontendError{
		Error:       "Bad vote request",
		Description: `Vote action specified in vote request has to be either "up" or "down"`,
	}
	UserNotFoundError = FrontendError{
		Error:       "User not found",
		Description: "No user with the specified ID exists.",
	}
	VoteOnSuggestedSongError = FrontendError{
		Error:       "User suggested song",
		Description: "The user requesting the vote has suggested this song.",
	}
	UserAlreadyVotedError = FrontendError{
		Error:       "User already voted",
		Description: "The user requesting the vote has already voted for this song.",
	}
	InternalServerError = FrontendError{
		Error:       "Internal server error",
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
