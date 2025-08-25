package grpcclient

import (
	"errors"
)

// Package level errors for the gRPC client layer
var (
	ErrorLogin    = errors.New("can't login")
	ErrorRegister = errors.New("can't register")
	ErrorDelete   = errors.New("can't delete data")
)
