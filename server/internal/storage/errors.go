package storage

import (
	"errors"
)

// Package level errors for the storage layer
var (
	ErrBadConn        = errors.New("error connecting to DB")
	ErrDuplicateLogin = errors.New("login already taken")
	ErrWrongPassword  = errors.New("login/password pair is wrong")
	ErrUnauthorized   = errors.New("user not logged in")
	ErrNoDataFound    = errors.New("no data found for user")
)
