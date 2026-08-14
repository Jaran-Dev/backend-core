package oauth

import (
	"context"
)

type Profile struct {
	ProviderUserID string
	Email          string
	DisplayName    string
	AvatarURL      string
	Locale         string
}

type OAuthProvider interface {
	Name() Provider
	AuthURL(state, codeChallenge string) string
	Exchange(ctx context.Context, code, codeVerifier string) (*Profile, error)
}

type StateStore interface {
	Save(ctx context.Context, state, codeVerifier, redirectURL string) error
	Consume(ctx context.Context, state string) (codeVerifier string, redirectURL string, err error)
}
