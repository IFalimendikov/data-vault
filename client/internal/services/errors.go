package services

import (
	"errors"
)

// Package level errors for the vault service layer
var (
	ErrorNotFound = errors.New("error finding data")
	ErrorNoDB     = errors.New("error connecting DB")
)
