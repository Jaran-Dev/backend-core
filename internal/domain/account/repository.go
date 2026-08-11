package account

import "context"

type Repository interface {
	Create(ctx context.Context, a *Account) error
	GetByEmail(ctx context.Context, email string) (*Account, error)
}
