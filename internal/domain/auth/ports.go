package auth

import (
	"context"
	"time"

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
	UserID    uuid.UUID
	Role      Role
	TokenType string // "access"
}

type TokenService interface {
	GenerateAccessToken(userID uuid.UUID, role Role) (string, time.Time, error)
	ValidateAccessToken(tokenString string) (*AccessTokenClaims, error)
}

type SessionStore interface {
	SaveSession(ctx context.Context, hashedToken string, userID uuid.UUID, role Role, familyID string, expiresAt time.Time) error
	GetSession(ctx context.Context, hashedToken string) (userID uuid.UUID, role Role, familyID string, used bool, expiresAt time.Time, err error)
	AtomicRotate(ctx context.Context, hashedOldToken, hashedNewToken string, newExpiresAt time.Time) (uuid.UUID, Role, error)
	RevokeFamily(ctx context.Context, familyID string) error
}
