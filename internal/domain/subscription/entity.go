package subscription

import (
	"time"

	"github.com/Jaran-Dev/backend-core/internal/domain/user"
	"github.com/google/uuid"
)

type Subscription struct {
	id          SubscriptionID
	userID      user.UserID
	source      Source
	startsAt    time.Time
	endsAt      time.Time // Zero value means lifetime access
	providerRef string    // Transaction ID from gateway (empty if manual)
	createdAt   time.Time
}

func NewSubscription(userID user.UserID, source Source, startsAt, endsAt time.Time, providerRef string) (*Subscription, error) {
	if !source.IsValid() {
		return nil, ErrInvalidSource
	}

	return &Subscription{
		id:          SubscriptionID(uuid.New()),
		userID:      userID,
		source:      source,
		startsAt:    startsAt,
		endsAt:      endsAt,
		providerRef: providerRef,
		createdAt:   time.Now(),
	}, nil
}

func (s *Subscription) ID() SubscriptionID {
	return s.id
}

func (s *Subscription) UserID() user.UserID {
	return s.userID
}

func (s *Subscription) Source() Source {
	return s.source
}

func (s *Subscription) StartsAt() time.Time {
	return s.startsAt
}

func (s *Subscription) EndsAt() time.Time {
	return s.endsAt
}

func (s *Subscription) ProviderRef() string {
	return s.providerRef
}

func (s *Subscription) CreatedAt() time.Time {
	return s.createdAt
}

func (s *Subscription) IsActive() bool {
	now := time.Now()
	if s.startsAt.After(now) {
		return false
	}
	if s.endsAt.IsZero() {
		return true
	}
	return s.endsAt.After(now)
}
