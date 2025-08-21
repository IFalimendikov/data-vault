package grpcclient

import (
	"errors"
)

// Package level errors for the URL shortener service layer
var (
	ErrorLogin    = errors.New("can't login")
	ErrorRegister = errors.New("can't login")
	ErrorDelete   = errors.New("can't delete data")
)
