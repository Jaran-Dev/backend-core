package auth

import "errors"

var (
	ErrInvalidCredentials = errors.New("invalid credentials")
	ErrTokenNotFound      = errors.New("refresh token not found")
	ErrTokenExpired       = errors.New("refresh token expired")
	ErrRefreshTokenReused = errors.New("refresh token reuse detected")
	ErrAccountNotVerified = errors.New("account not verified")
)
