package user

import (
	"context"
	"errors"
	"time"

	domainAccount "github.com/Jaran-Dev/backend-core/internal/domain/account"
	domainUser "github.com/Jaran-Dev/backend-core/internal/domain/user"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

// We define both models here to handle the atomic transaction cleanly
type AccountModel struct {
	ID           uuid.UUID `gorm:"primaryKey"`
	Email        string    `gorm:"uniqueIndex"`
	PasswordHash string
	Role         string
	Status       string
	CreatedAt    time.Time
	UpdatedAt    time.Time
}

func (AccountModel) TableName() string { return "accounts" }

type UserModel struct {
	AccountID   uuid.UUID `gorm:"primaryKey"`
	DisplayName string
	UILocale    string
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

func (UserModel) TableName() string { return "users" }

type pgRepository struct {
	db *gorm.DB
}

func NewRepository(db *gorm.DB) domainUser.Repository {
	return &pgRepository{db: db}
}

func (r *pgRepository) CreateWithAccount(ctx context.Context, acc *domainAccount.Account, u *domainUser.User) error {
	accModel := AccountModel{
		ID:           uuid.UUID(acc.ID()),
		Email:        acc.Email(),
		PasswordHash: acc.PasswordHash(),
		Role:         string(acc.Role()),
		Status:       string(acc.Status()),
		CreatedAt:    acc.CreatedAt(),
		UpdatedAt:    acc.UpdatedAt(),
	}

	userModel := UserModel{
		AccountID:   uuid.UUID(u.AccountID()),
		DisplayName: u.DisplayName(),
		UILocale:    u.UILocale(),
		CreatedAt:   acc.CreatedAt(),
		UpdatedAt:   acc.UpdatedAt(),
	}

	err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(&accModel).Error; err != nil {
			if errors.Is(err, gorm.ErrDuplicatedKey) {
				return domainAccount.ErrAccountAlreadyExists
			}
			return err
		}

		if err := tx.Create(&userModel).Error; err != nil {
			return err
		}

		return nil
	})

	return err
}

func (r *pgRepository) GetByAccountID(ctx context.Context, accID domainAccount.AccountID) (*domainUser.User, error) {
	var model UserModel
	err := r.db.WithContext(ctx).First(&model, "account_id = ?", uuid.UUID(accID)).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("user profile not found")
		}
		return nil, err
	}

	return domainUser.UnmarshalUser(domainAccount.AccountID(model.AccountID), model.DisplayName, model.UILocale, model.CreatedAt, model.UpdatedAt), nil
}
