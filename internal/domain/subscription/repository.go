package subscription

import (
	"context"

	"github.com/Jaran-Dev/backend-core/internal/domain/account"
)

type Repository interface {
	Create(ctx context.Context, s *Subscription) error
	GetActiveByUserID(ctx context.Context, accountID account.AccountID) (*Subscription, error)
	ListByUserID(ctx context.Context, accountID account.AccountID) ([]*Subscription, error)
}
