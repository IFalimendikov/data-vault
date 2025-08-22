package grpcclient

import (
	"errors"
)

// Package level errors for the gRPC client layer
var (
	ErrorLogin    = errors.New("can't login")       // Login operation failed
	ErrorRegister = errors.New("can't register")    // Registration operation failed
	ErrorDelete   = errors.New("can't delete data") // Data deletion failed
)
