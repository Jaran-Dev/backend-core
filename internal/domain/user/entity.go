package user

import (
	"errors"
	"time"

	"github.com/google/uuid"
)

type User struct {
	id           UserID
	email        Email
	passwordHash string
	displayName  string
	uiLocale     string // "en", "ckb", "kmr"
	createdAt    time.Time
}

func NewUser(email Email, passwordHash, displayName, uiLocale string) *User {
	return &User{
		id:           UserID(uuid.New()),
		email:        email,
		passwordHash: passwordHash,
		displayName:  displayName,
		uiLocale:     uiLocale,
		createdAt:    time.Now(),
	}
}

func (u *User) ID() UserID {
	return u.id
}

func (u *User) Email() Email {
	return u.email
}

func (u *User) PasswordHash() string {
	return u.passwordHash
}

func (u *User) DisplayName() string {
	return u.displayName
}

func (u *User) UILocale() string {
	return u.uiLocale
}

func (u *User) CreatedAt() time.Time {
	return u.createdAt
}

func (u *User) UpdateProfile(displayName, uiLocale string) {
	if displayName != "" {
		u.displayName = displayName
	}
	if uiLocale != "" {
		u.uiLocale = uiLocale
	}
}

func UnmarshalUser(id uuid.UUID, emailStr string, passwordHash string, displayName *string, uiLocale string, createdAt time.Time) (*User, error) {
	if passwordHash == "" {
		return nil, errors.New("cannot unmarshal user: password hash is empty")
	}

	email, err := NewEmail(emailStr)
	if err != nil {
		return nil, errors.New("cannot unmarshal user: invalid email in database")
	}

	name := ""
	if displayName != nil {
		name = *displayName
	}

	return &User{
		id:           UserID(id),
		email:        email,
		passwordHash: passwordHash,
		displayName:  name,
		uiLocale:     uiLocale,
		createdAt:    createdAt,
	}, nil
}
