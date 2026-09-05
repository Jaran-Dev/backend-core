package admin

import (
	"context"
	"errors"
	"time"

	domainAccount "github.com/Jaran-Dev/backend-core/internal/domain/account"
	domainAdmin "github.com/Jaran-Dev/backend-core/internal/domain/admin"
	"github.com/Jaran-Dev/backend-core/internal/infrastructure/repository"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type AdminORM struct {
	AccountID uuid.UUID `gorm:"primaryKey"`
	CreatedAt time.Time
}

func (AdminORM) TableName() string { return "admins" }

func (m AdminORM) toDomain() *domainAdmin.Admin {
	return domainAdmin.UnmarshalAdmin(domainAccount.AccountID(m.AccountID), m.CreatedAt)
}

func toPersistence(a *domainAdmin.Admin) AdminORM {
	return AdminORM{
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
	orm := toPersistence(a)
	db := repository.GetDB(ctx, r.db)

	return db.Create(&orm).Error
}

func (r *pgRepository) GetByAccountID(ctx context.Context, accID domainAccount.AccountID) (*domainAdmin.Admin, error) {
	var orm AdminORM
	db := repository.GetDB(ctx, r.db)

	err := db.First(&orm, "account_id = ?", uuid.UUID(accID)).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("admin profile not found")
		}
		return nil, err
	}
	return orm.toDomain(), nil
}
