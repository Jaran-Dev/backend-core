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
	GenerateEmailVerificationToken(accountID account.AccountID) (string, error)
	ValidateEmailVerificationToken(tokenString string) (account.AccountID, error)
}

type SessionStore interface {
	SaveSession(ctx context.Context, hashedToken string, accountID account.AccountID, role Role, familyID string, expiresAt time.Time) error
	GetSession(ctx context.Context, hashedToken string) (accountID account.AccountID, role Role, familyID string, used bool, expiresAt time.Time, err error)
	AtomicRotate(ctx context.Context, hashedOldToken, hashedNewToken string, newExpiresAt time.Time) (account.AccountID, Role, error)
	RevokeFamily(ctx context.Context, familyID string) error
	RevokeAllByAccountID(ctx context.Context, accountID account.AccountID) error
	SavePasswordResetToken(ctx context.Context, accountID account.AccountID, hashedToken string, expiresAt time.Time) error
	GetAccountIDByPasswordResetToken(ctx context.Context, hashedToken string) (account.AccountID, error)
	DeletePasswordResetToken(ctx context.Context, hashedToken string) error
}

type EmailService interface {
	SendVerificationEmail(to, token string) error
	SendPasswordResetEmail(to, token string) error
}

type LoginAttemptStore interface {
	IncrementFailedAttempt(ctx context.Context, email string) (int, error)
	ResetFailedAttempts(ctx context.Context, email string) error
	LockAccount(ctx context.Context, accountID account.AccountID, duration time.Duration) error
	IsAccountLocked(ctx context.Context, accountID account.AccountID) (bool, error)
	UnlockAccount(ctx context.Context, accountID account.AccountID) error
}
