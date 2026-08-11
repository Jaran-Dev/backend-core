package admin

import (
	"context"
	"errors"
	"time"

	domainAccount "github.com/Jaran-Dev/backend-core/internal/domain/account"
	domainAdmin "github.com/Jaran-Dev/backend-core/internal/domain/admin"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type AdminModel struct {
	AccountID uuid.UUID `gorm:"primaryKey"`
	CreatedAt time.Time
}

func (AdminModel) TableName() string { return "admins" }

func (m AdminModel) toDomain() *domainAdmin.Admin {
	return domainAdmin.UnmarshalAdmin(domainAccount.AccountID(m.AccountID), m.CreatedAt)
}

func adminModelFromDomain(a *domainAdmin.Admin) AdminModel {
	return AdminModel{
		AccountID: uuid.UUID(a.AccountID()),
		CreatedAt: a.CreatedAt(),
	}
}

type pgRepository struct {
	db *gorm.DB
}

func NewRepository(db *gorm.DB) domainAdmin.Repository {
	return &pgRepository{db: db}
}

func (r *pgRepository) Create(ctx context.Context, a *domainAdmin.Admin) error {
	model := adminModelFromDomain(a)
	return r.db.WithContext(ctx).Create(&model).Error
}

func (r *pgRepository) GetByAccountID(ctx context.Context, accID domainAccount.AccountID) (*domainAdmin.Admin, error) {
	var model AdminModel
	err := r.db.WithContext(ctx).First(&model, "account_id = ?", uuid.UUID(accID)).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("admin profile not found")
		}
		return nil, err
	}
	return model.toDomain(), nil
}
