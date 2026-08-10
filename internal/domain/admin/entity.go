package admin

import (
	"errors"
	"time"

	"github.com/google/uuid"
)

type Admin struct {
	id           AdminID
	email        Email
	passwordHash string
	createdAt    time.Time
}

func NewAdmin(email Email, passwordHash string) *Admin {
	return &Admin{
		id:           AdminID(uuid.New()),
		email:        email,
		passwordHash: passwordHash,
		createdAt:    time.Now(),
	}
}

func (a *Admin) ID() AdminID {
	return a.id
}

func (a *Admin) Email() Email {
	return a.email
}

func (a *Admin) PasswordHash() string {
	return a.passwordHash
}

func (a *Admin) CreatedAt() time.Time {
	return a.createdAt
}

func (a *Admin) UpdatePassword(newHash string) {
	if newHash != "" {
		a.passwordHash = newHash
	}
}

func UnmarshalAdmin(id uuid.UUID, emailStr string, passwordHash string, createdAt time.Time) (*Admin, error) {
	if passwordHash == "" {
		return nil, errors.New("cannot unmarshal admin: password hash is empty")
	}

	email, err := NewEmail(emailStr)
	if err != nil {
		return nil, errors.New("cannot unmarshal admin: invalid email in database")
	}

	return &Admin{
		id:           AdminID(id),
		email:        email,
		passwordHash: passwordHash,
		createdAt:    createdAt,
	}, nil
}
