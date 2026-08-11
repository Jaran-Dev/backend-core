package account

import (
	"errors"
	"time"

	"github.com/google/uuid"
)

type AccountID uuid.UUID

func (id AccountID) String() string { return uuid.UUID(id).String() }

type Role string

const (
	RoleUser  Role = "user"
	RoleAdmin Role = "admin"
)

type Status string

const (
	StatusActive              Status = "active"
	StatusBanned              Status = "banned"
	StatusLocked              Status = "locked"
	StatusPendingVerification Status = "pending_verification"
)

type Account struct {
	id           AccountID
	email        string
	passwordHash string
	role         Role
	status       Status
	createdAt    time.Time
	updatedAt    time.Time
}

func NewAccount(email, passwordHash string, role Role) (*Account, error) {
	if role != RoleUser && role != RoleAdmin {
		return nil, errors.New("invalid role")
	}
	return &Account{
		id:           AccountID(uuid.New()),
		email:        email,
		passwordHash: passwordHash,
		role:         role,
		// Default to pending verification. Admins created manually could bypass this.
		status:    StatusPendingVerification,
		createdAt: time.Now(),
		updatedAt: time.Now(),
	}, nil
}

func (a *Account) ID() AccountID        { return a.id }
func (a *Account) Email() string        { return a.email }
func (a *Account) PasswordHash() string { return a.passwordHash }
func (a *Account) Role() Role           { return a.role }
func (a *Account) Status() Status       { return a.status }
func (a *Account) CreatedAt() time.Time { return a.createdAt }
func (a *Account) UpdatedAt() time.Time { return a.updatedAt }

func (a *Account) Verify() {
	if a.status == StatusPendingVerification {
		a.status = StatusActive
		a.updatedAt = time.Now()
	}
}

func (a *Account) UpdatePasswordHash(newHash string) {
	a.passwordHash = newHash
	a.updatedAt = time.Now()
}

func UnmarshalAccount(id uuid.UUID, email, passwordHash, roleStr, statusStr string, createdAt, updatedAt time.Time) (*Account, error) {
	role := Role(roleStr)
	if role != RoleUser && role != RoleAdmin {
		return nil, errors.New("invalid role in database")
	}
	return &Account{
		id:           AccountID(id),
		email:        email,
		passwordHash: passwordHash,
		role:         role,
		status:       Status(statusStr),
		createdAt:    createdAt,
		updatedAt:    updatedAt,
	}, nil
}
