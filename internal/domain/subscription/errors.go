package subscription

import "errors"

var (
	ErrSubscriptionNotFound = errors.New("active subscription not found")
	ErrInvalidSource        = errors.New("invalid subscription source")
	ErrAlreadySubscribed    = errors.New("user already has an active subscription")
)
