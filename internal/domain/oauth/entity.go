package oauth

import (
	"time"

	"github.com/Jaran-Dev/backend-core/internal/domain/account"
	"github.com/google/uuid"
)

type OAuthAccount struct {
	id             OAuthAccountID
	accountID      account.AccountID
	provider       Provider
	providerUserID string
	emailSnapshot  string
	createdAt      time.Time
	updatedAt      time.Time
}

func NewOAuthAccount(accID account.AccountID, provider Provider, providerUserID, emailSnapshot string) *OAuthAccount {
	now := time.Now()
	return &OAuthAccount{
		id:             OAuthAccountID(uuid.New()),
		accountID:      accID,
		provider:       provider,
		providerUserID: providerUserID,
		emailSnapshot:  emailSnapshot,
		createdAt:      now,
		updatedAt:      now,
	}
}

func UnmarshalOAuthAccount(
	id uuid.UUID,
	accID account.AccountID,
	provider Provider,
	providerUserID, emailSnapshot string,
	createdAt, updatedAt time.Time,
) (*OAuthAccount, error) {
	if !provider.IsValid() {
		return nil, ErrOAuthProviderUnsupported
	}
	return &OAuthAccount{
		id:             OAuthAccountID(id),
		accountID:      accID,
		provider:       provider,
		providerUserID: providerUserID,
		emailSnapshot:  emailSnapshot,
		createdAt:      createdAt,
		updatedAt:      updatedAt,
	}, nil
}

func (o *OAuthAccount) ID() OAuthAccountID           { return o.id }
func (o *OAuthAccount) AccountID() account.AccountID { return o.accountID }
func (o *OAuthAccount) Provider() Provider           { return o.provider }
func (o *OAuthAccount) ProviderUserID() string       { return o.providerUserID }
func (o *OAuthAccount) EmailSnapshot() string        { return o.emailSnapshot }
func (o *OAuthAccount) CreatedAt() time.Time         { return o.createdAt }
func (o *OAuthAccount) UpdatedAt() time.Time         { return o.updatedAt }
