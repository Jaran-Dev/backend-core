package user

import (
	"context"

	"github.com/Jaran-Dev/backend-core/internal/domain/account"
)

type Repository interface {
	CreateWithAccount(ctx context.Context, acc *account.Account, u *User) error
	GetByAccountID(ctx context.Context, accID account.AccountID) (*User, error)
}
