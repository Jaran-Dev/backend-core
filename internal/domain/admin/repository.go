package admin

import (
	"context"

	"github.com/Jaran-Dev/backend-core/internal/domain/account"
)

type Repository interface {
	Create(ctx context.Context, a *Admin) error
	GetByAccountID(ctx context.Context, accID account.AccountID) (*Admin, error)
}
