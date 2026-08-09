package subscription

import (
	"context"

	"github.com/Jaran-Dev/backend-core/internal/domain/user"
)

type Repository interface {
	Create(ctx context.Context, s *Subscription) error
	GetActiveByUserID(ctx context.Context, userID user.UserID) (*Subscription, error)
	ListByUserID(ctx context.Context, userID user.UserID) ([]*Subscription, error)
}
