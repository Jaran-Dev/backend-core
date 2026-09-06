package preferences

import (
	"context"

	"github.com/Jaran-Dev/backend-core/internal/domain/account"
)

type Repository interface {
	Create(ctx context.Context, p *Preferences) error
	GetByAccountID(ctx context.Context, accID account.AccountID) (*Preferences, error)
	Update(ctx context.Context, p *Preferences) error
}
