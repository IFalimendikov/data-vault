package service

import (
	"errors"
)

var (
	ErrWrongFormat = errors.New("order number is in the wrong format")
	ErrNoNewAddresses = errors.New("no new addresses found")
	ErrMalformedRequest = errors.New("malformed request")
)