package shared

import "errors"

var (
	ErrInvalidLocale = errors.New("invalid locale: must be en, ckb, or kmr")
)
