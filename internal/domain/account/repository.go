package account

import "context"

type Repository interface {
	GetByEmail(ctx context.Context, email string) (*Account, error)
	GetByID(ctx context.Context, id AccountID) (*Account, error)
	Update(ctx context.Context, a *Account) error
}
