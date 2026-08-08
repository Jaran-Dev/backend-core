package admin

import "context"

type AdminRepository interface {
	Create(ctx context.Context, a *Admin) error
	GetByID(ctx context.Context, id AdminID) (*Admin, error)
	GetByEmail(ctx context.Context, email Email) (*Admin, error)
}
