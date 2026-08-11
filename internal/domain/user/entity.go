package user

import (
	"time"

	"github.com/Jaran-Dev/backend-core/internal/domain/account"
)

type User struct {
	accountID   account.AccountID
	displayName string
	uiLocale    string
	createdAt   time.Time
	updatedAt   time.Time
}

func NewUser(accID account.AccountID, displayName, uiLocale string) *User {
	return &User{
		accountID:   accID,
		displayName: displayName,
		uiLocale:    uiLocale,
		createdAt:   time.Now(),
		updatedAt:   time.Now(),
	}
}

func (u *User) AccountID() account.AccountID { return u.accountID }
func (u *User) DisplayName() string          { return u.displayName }
func (u *User) UILocale() string             { return u.uiLocale }

func UnmarshalUser(accID account.AccountID, displayName, uiLocale string, createdAt, updatedAt time.Time) *User {
	return &User{
		accountID:   accID,
		displayName: displayName,
		uiLocale:    uiLocale,
		createdAt:   createdAt,
		updatedAt:   updatedAt,
	}
}
