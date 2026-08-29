package user

import "errors"

var (
	ErrNotFound     = errors.New("user not found")
	ErrEmailExists  = errors.New("email already registered")
	ErrInvalidCreds = errors.New("invalid email or password")
)
