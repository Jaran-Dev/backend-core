package account

import (
	"context"
	"errors"
	"time"

	domainAccount "github.com/Jaran-Dev/backend-core/internal/domain/account"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

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

type pgRepository struct {
	db *gorm.DB
}

func NewRepository(db *gorm.DB) domainAccount.Repository {
	return &pgRepository{db: db}
}

func (r *pgRepository) GetByEmail(ctx context.Context, email string) (*domainAccount.Account, error) {
	var model AccountModel
	err := r.db.WithContext(ctx).Where("email = ?", email).First(&model).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, domainAccount.ErrAccountNotFound
		}
		return nil, err
	}

	return domainAccount.UnmarshalAccount(
		model.ID, model.Email, model.PasswordHash, model.Role, model.Status, model.CreatedAt, model.UpdatedAt,
	)
}

func (r *pgRepository) GetByID(ctx context.Context, id domainAccount.AccountID) (*domainAccount.Account, error) {
	var model AccountModel
	err := r.db.WithContext(ctx).First(&model, "id = ?", uuid.UUID(id)).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, domainAccount.ErrAccountNotFound
		}
		return nil, err
	}

	return domainAccount.UnmarshalAccount(
		model.ID, model.Email, model.PasswordHash, model.Role, model.Status, model.CreatedAt, model.UpdatedAt,
	)
}

func (r *pgRepository) Update(ctx context.Context, acc *domainAccount.Account) error {
	model := AccountModel{
		ID:           uuid.UUID(acc.ID()),
		Email:        acc.Email(),
		PasswordHash: acc.PasswordHash(),
		Role:         string(acc.Role()),
		Status:       string(acc.Status()),
		CreatedAt:    acc.CreatedAt(),
		UpdatedAt:    acc.UpdatedAt(),
	}

	return r.db.WithContext(ctx).Save(&model).Error
}
