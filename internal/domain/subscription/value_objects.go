package subscription

import "github.com/google/uuid"

type SubscriptionID uuid.UUID

func (id SubscriptionID) String() string {
	return uuid.UUID(id).String()
}

type Source string

const (
	SourceManual  Source = "manual"  // Admin granted it manually
	SourceGateway Source = "gateway" // Paid via local Kurdish gateway/crypto (future)
	SourcePromo   Source = "promo"   // Promotional code
)

func (s Source) IsValid() bool {
	switch s {
	case SourceManual, SourceGateway, SourcePromo:
		return true
	}
	return false
}
