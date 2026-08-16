package module

import "errors"

var (
	ErrModuleNotFound         = errors.New("module not found")
	ErrInvalidPosition        = errors.New("invalid position: must be 0 or greater")
	ErrConcurrentModification = errors.New("concurrent modification detected: the data was updated by another transaction")
	ErrInvalidModuleTitle     = errors.New("invalid module title: cannot be empty")
	ErrInvalidReorderPayload  = errors.New("invalid reorder payload")
)
