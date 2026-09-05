package account

import (
	"context"
	"errors"
	"time"

	domainAccount "github.com/Jaran-Dev/backend-core/internal/domain/account"
	"github.com/Jaran-Dev/backend-core/internal/infrastructure/repository"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type AccountORM struct {
	ID           uuid.UUID `gorm:"primaryKey"`
	Email        string    `gorm:"uniqueIndex"`
	PasswordHash *string   `gorm:"column:password_hash"`
	Role         string
	Status       string
	CreatedAt    time.Time
	UpdatedAt    time.Time
}

func (AccountORM) TableName() string { return "accounts" }

func (m AccountORM) toDomain() (*domainAccount.Account, error) {
	return domainAccount.UnmarshalAccount(
		m.ID, m.Email, m.PasswordHash, m.Role, m.Status, m.CreatedAt, m.UpdatedAt,
	)
}

func ToPersistence(acc *domainAccount.Account) AccountORM {
	return AccountORM{
		ID:           uuid.UUID(acc.ID()),
		Email:        acc.Email(),
		PasswordHash: acc.PasswordHash(),
		Role:         string(acc.Role()),
		Status:       string(acc.Status()),
		CreatedAt:    acc.CreatedAt(),
		UpdatedAt:    acc.UpdatedAt(),
	}
}

type pgRepository struct {
	db *gorm.DB
}

func NewRepository(db *gorm.DB) domainAccount.Repository {
	return &pgRepository{db: db}
}

func (r *pgRepository) Create(ctx context.Context, acc *domainAccount.Account) error {
	orm := ToPersistence(acc)
	db := repository.GetDB(ctx, r.db)

	if err := db.Create(&orm).Error; err != nil {
		translatedErr := repository.TranslateGormError(err)
		if errors.Is(translatedErr, gorm.ErrDuplicatedKey) {
			return domainAccount.ErrAccountAlreadyExists
		}
		return translatedErr
	}
	return nil
}

func (r *pgRepository) GetByEmail(ctx context.Context, email string) (*domainAccount.Account, error) {
	var orm AccountORM
	db := repository.GetDB(ctx, r.db)

	err := db.Where("email = ?", email).First(&orm).Error
	if err != nil {
		translatedErr := repository.TranslateGormError(err)
		if errors.Is(translatedErr, gorm.ErrRecordNotFound) {
			return nil, domainAccount.ErrAccountNotFound
		}
		return nil, translatedErr
	}
	return orm.toDomain()
}

func (r *pgRepository) GetByID(ctx context.Context, id domainAccount.AccountID) (*domainAccount.Account, error) {
	var orm AccountORM
	db := repository.GetDB(ctx, r.db)

	err := db.First(&orm, "id = ?", uuid.UUID(id)).Error
	if err != nil {
		translatedErr := repository.TranslateGormError(err)
		if errors.Is(translatedErr, gorm.ErrRecordNotFound) {
			return nil, domainAccount.ErrAccountNotFound
		}
		return nil, translatedErr
	}
	return orm.toDomain()
}

func (r *pgRepository) Update(ctx context.Context, acc *domainAccount.Account) error {
	orm := ToPersistence(acc)
	db := repository.GetDB(ctx, r.db)

	if err := db.Save(&orm).Error; err != nil {
		return repository.TranslateGormError(err)
	}
	return nil
}
