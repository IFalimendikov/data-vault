package storage

import (
	"errors"
)

var (
	ErrBadConn = errors.New("error connecting to DB")
	ErrDuplicateLogin = errors.New("login already taken")
	ErrDuplicateNumber = errors.New("order uploaded by different user")
	ErrDuplicateOrder = errors.New("order already received")
	ErrWrongPassword = errors.New("login/password pair is wrong")
	ErrUnauthorized = errors.New("user not logged in")
	ErrNoOrdersFound = errors.New("client has no orders")
	ErrNoWithdrawalsFound = errors.New("client has no withdrawals")
	ErrBalanceTooLow = errors.New("balance is too low for the transcation")
)