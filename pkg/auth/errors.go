package auth

import (
	"errors"
)

var (
	ErrInvalidAuthToken = errors.New("invalid-token")
	ErrExpiredAuthToken = errors.New("token-expired")
	ErrNoAuthToken      = errors.New("no-auth-token")
)
