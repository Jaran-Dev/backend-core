package admin

import "errors"

var (
	ErrAdminNotFound      = errors.New("admin not found")
	ErrInvalidEmail       = errors.New("invalid email format")
	ErrInvalidCredential  = errors.New("invalid credentials")
	ErrAdminAlreadyExists = errors.New("admin with this email already exists")
)
