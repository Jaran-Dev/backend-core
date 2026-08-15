package oauth

import (
	"context"

	"github.com/Jaran-Dev/backend-core/internal/domain/account"
)

type Repository interface {
	GetByProvider(ctx context.Context, p Provider, puid string) (*OAuthAccount, error)
	GetByAccountID(ctx context.Context, accID account.AccountID) ([]*OAuthAccount, error)
	Create(ctx context.Context, o *OAuthAccount) error
	Delete(ctx context.Context, id OAuthAccountID) error
}
