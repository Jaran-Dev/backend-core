package account

import "context"

type Repository interface {
	GetByEmail(ctx context.Context, email string) (*Account, error)
}
