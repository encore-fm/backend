package handlers

import (
	"errors"
	log "github.com/sirupsen/logrus"
	"net/http"
)

type FrontendError struct {
	error       string
	description string
}

var (
	// Spotify errors
	ErrSpotifyNotAuthenticated = errors.New("spotify not authenticated")
	// Vote errors
	ErrBadVoteAction = errors.New(`vote action must be in {"up", "down"}`)

	// Frontend errors
	SessionNotFoundError = FrontendError{
		error:       "Session not found",
		description: "No session with the specified ID exists.",
	}
	UserConflictError = FrontendError{
		error:       "Username already exists",
		description: "A user with the given username already exists.",
	}
	SongConflictError = FrontendError{
		error:       "Song already suggested",
		description: "The song being requested has already been suggested in this session.",
	}
	BadVoteError = FrontendError{
		error:       "Bad vote request",
		description: `Vote action specified in vote request has to be either "up" or "down"`,
	}
	UserNotFoundError = FrontendError{
		error:       "User not found",
		description: "No user with the specified ID exists.",
	}
	VoteOnSuggestedSongError = FrontendError{
		error:       "User suggested song",
		description: "The user requesting the vote has suggested this song.",
	}
	UserAlreadyVotedError = FrontendError{
		error:       "User already voted",
		description: "The user requesting the vote has already voted for this song.",
	}
	InternalServerError = FrontendError{
		error:       "Internal server error",
		description: "An unexpected server error has occurred.",
	}
)

func HandleError(w http.ResponseWriter, status int, logLevel log.Level, msg string, err error, frontendError FrontendError) {
	switch logLevel {
	case log.WarnLevel:
		log.Warnf(msg, err)
		break
	case log.ErrorLevel:
		log.Errorf(msg, err)
		break
	default:
		log.Infof(msg, err)
		break
	}

	jsonResponseWithStatus(w, status, frontendError)
}
