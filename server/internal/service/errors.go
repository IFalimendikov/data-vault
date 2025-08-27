package service

import (
	"errors"
)

// Service layer error definitions
var (
	ErrWrongFormat      = errors.New("order number is in the wrong format")
	ErrNoNewAddresses   = errors.New("no new addresses found")
	ErrMalformedRequest = errors.New("malformed request")
)
