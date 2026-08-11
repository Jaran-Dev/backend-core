package user

import (
	"context"

	"github.com/Jaran-Dev/backend-core/internal/domain/account"
)

type Repository interface {
	Create(ctx context.Context, u *User) error
	GetByAccountID(ctx context.Context, accID account.AccountID) (*User, error)
}
