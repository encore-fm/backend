package db

import "errors"

var (
	ErrIllegalState = errors.New("illegal state in database")

	// User collection errors
	ErrUsernameTaken   = errors.New("requested username already taken")
	ErrNoUserWithID    = errors.New("no user with given id")
	ErrNoUserWithState = errors.New("no user with given state")

	// Song collection errors
	ErrNoSongWithID         = errors.New("no song with given id")
	ErrUserAlreadyVoted     = errors.New("user already voted for this song")
	ErrSongAlreadySuggested = errors.New("song has already been suggested")

	// Session collection errors
	ErrSessionAlreadyExisting = errors.New("session with this id already exists")
	ErrNoSessionWithID        = errors.New("no session with given id")
	ErrSongAlreadyInSession   = errors.New("song with this ID already exists for this session")
)
