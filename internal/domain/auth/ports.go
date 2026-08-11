package auth

import (
	"context"
	"time"

	"github.com/Jaran-Dev/backend-core/internal/domain/account"
)

type Role string

const (
	RoleUser  Role = "user"
	RoleAdmin Role = "admin"
)

type Hasher interface {
	Hash(password string) (string, error)
	Compare(hash, password string) (bool, error)
}

type AccessTokenClaims struct {
	AccountID account.AccountID
	Role      Role
}

type TokenService interface {
	GenerateAccessToken(accountID account.AccountID, role Role) (string, time.Time, error)
	ValidateAccessToken(tokenString string) (*AccessTokenClaims, error)
}

type SessionStore interface {
	SaveSession(ctx context.Context, hashedToken string, accountID account.AccountID, role Role, familyID string, expiresAt time.Time) error
	GetSession(ctx context.Context, hashedToken string) (accountID account.AccountID, role Role, familyID string, used bool, expiresAt time.Time, err error)
	AtomicRotate(ctx context.Context, hashedOldToken, hashedNewToken string, newExpiresAt time.Time) (account.AccountID, Role, error)
	RevokeFamily(ctx context.Context, familyID string) error
}
