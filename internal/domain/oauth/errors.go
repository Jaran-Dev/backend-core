package oauth

import "errors"

var (
	ErrOAuthAccountNotFound     = errors.New("oauth account not found")
	ErrAccountAlreadyLinked     = errors.New("account is already linked to this provider")
	ErrInvalidState             = errors.New("invalid or expired oauth state")
	ErrOAuthProviderUnsupported = errors.New("unsupported oauth provider")
	ErrOAuthCodeExchangeFailed  = errors.New("failed to exchange oauth code")
	ErrOAuthProfileFetchFailed  = errors.New("failed to fetch oauth profile")
	ErrCannotUnlinkLastMethod   = errors.New("cannot unlink the last authentication method")
)
