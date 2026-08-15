package auth

import (
	"context"
	"time"

	"github.com/Jaran-Dev/backend-core/internal/domain/account"
	"github.com/google/uuid"
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

type SessionMetadata struct {
	IPAddress  string
	UserAgent  string
	DeviceType string
}

type Session struct {
	ID           uuid.UUID
	AccountID    account.AccountID
	Role         Role
	ExpiresAt    time.Time
	LastActiveAt time.Time
	CreatedAt    time.Time
	Metadata     SessionMetadata
}

type SessionStore interface {
	SaveSession(ctx context.Context, sessionID uuid.UUID, hashedToken string, accID account.AccountID, role Role, expiresAt time.Time, meta SessionMetadata) error
	GetSession(ctx context.Context, hashedToken string) (sessionID uuid.UUID, accID account.AccountID, role Role, used bool, expiresAt time.Time, err error)
	AtomicRotate(ctx context.Context, hashedOldToken, hashedNewToken string, newExpiresAt time.Time, meta SessionMetadata) (uuid.UUID, account.AccountID, Role, error)
	RevokeSessionByID(ctx context.Context, accountID account.AccountID, sessionID uuid.UUID) error
	ListSessionsByAccountID(ctx context.Context, accountID account.AccountID) ([]*Session, error)
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
