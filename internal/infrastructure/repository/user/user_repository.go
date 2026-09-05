package user

import (
	"context"
	"errors"
	"time"

	domainAccount "github.com/Jaran-Dev/backend-core/internal/domain/account"
	domainUser "github.com/Jaran-Dev/backend-core/internal/domain/user"
	repoAccount "github.com/Jaran-Dev/backend-core/internal/infrastructure/repository/account"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type UserORM struct {
	AccountID   uuid.UUID `gorm:"primaryKey"`
	DisplayName string
	UILocale    string
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

func (UserORM) TableName() string { return "users" }

func (m UserORM) toDomain() *domainUser.User {
	return domainUser.UnmarshalUser(
		domainAccount.AccountID(m.AccountID), m.DisplayName, m.UILocale, m.CreatedAt, m.UpdatedAt,
	)
}

func toPersistence(u *domainUser.User) UserORM {
	return UserORM{
		AccountID:   uuid.UUID(u.AccountID()),
		DisplayName: u.DisplayName(),
		UILocale:    u.UILocale(),
		CreatedAt:   u.CreatedAt(),
		UpdatedAt:   u.UpdatedAt(),
	}
}

type pgRepository struct {
	db *gorm.DB
}

func NewRepository(db *gorm.DB) domainUser.Repository {
	return &pgRepository{db: db}
}

func (r *pgRepository) CreateWithAccount(ctx context.Context, acc *domainAccount.Account, u *domainUser.User) error {
	accORM := repoAccount.ToPersistence(acc)
	userORM := toPersistence(u)

	err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(&accORM).Error; err != nil {
			if errors.Is(err, gorm.ErrDuplicatedKey) {
				return domainAccount.ErrAccountAlreadyExists
			}
			return err
		}

		if err := tx.Create(&userORM).Error; err != nil {
			return err
		}

		return nil
	})

	return err
}

func (r *pgRepository) GetByAccountID(ctx context.Context, accID domainAccount.AccountID) (*domainUser.User, error) {
	var orm UserORM
	err := r.db.WithContext(ctx).First(&orm, "account_id = ?", uuid.UUID(accID)).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("user profile not found")
		}
		return nil, err
	}

	return orm.toDomain(), nil
}
