package account

import "errors"

var (
	ErrAccountNotFound      = errors.New("account not found")
	ErrAccountAlreadyExists = errors.New("account with this email already exists")
	ErrAccountBanned        = errors.New("account is banned")
)
