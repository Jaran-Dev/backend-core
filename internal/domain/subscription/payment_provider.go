package subscription

import (
	"context"

	"github.com/Jaran-Dev/backend-core/internal/domain/user"
)

type CheckoutResult struct {
	CheckoutURL string
	ProviderRef string
}

type WebhookResult struct {
	UserID      user.UserID
	ProviderRef string
	Source      Source
	StartsAt    interface{}
	EndsAt      interface{}
}

// PaymentProvider is the PORT for external payment gateways.
// The Application Service will depend on this interface, not a concrete implementation.
// For MVP, we will build a ManualProvider in the infrastructure layer.
// In v1.1, we will build a FastPayProvider or CryptoProvider that implements this exact interface.
type PaymentProvider interface {
	Name() string
	InitiateSubscription(ctx context.Context, u *user.User) (CheckoutResult, error)
	HandleWebhook(ctx context.Context, payload []byte, signature string) (WebhookResult, error)
}
