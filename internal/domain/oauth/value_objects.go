package oauth

import "github.com/google/uuid"

type OAuthAccountID uuid.UUID

func (id OAuthAccountID) String() string { return uuid.UUID(id).String() }

type Provider string

const (
	ProviderGoogle Provider = "google"
	ProviderGitHub Provider = "github"
	ProviderApple  Provider = "apple"
)

func (p Provider) IsValid() bool {
	switch p {
	case ProviderGoogle, ProviderGitHub, ProviderApple:
		return true
	}
	return false
}
