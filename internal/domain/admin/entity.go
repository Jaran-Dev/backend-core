package admin

import (
	"time"

	"github.com/Jaran-Dev/backend-core/internal/domain/account"
)

type Admin struct {
	accountID account.AccountID
	createdAt time.Time
}

func NewAdmin(accID account.AccountID) *Admin {
	return &Admin{
		accountID: accID,
		createdAt: time.Now(),
	}
}

func (a *Admin) AccountID() account.AccountID { return a.accountID }
func (a *Admin) CreatedAt() time.Time         { return a.createdAt }

func UnmarshalAdmin(accID account.AccountID, createdAt time.Time) *Admin {
	return &Admin{
		accountID: accID,
		createdAt: createdAt,
	}
}
