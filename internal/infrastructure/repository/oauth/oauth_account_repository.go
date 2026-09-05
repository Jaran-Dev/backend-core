package oauth

import (
	"context"
	"errors"
	"time"

	domainAccount "github.com/Jaran-Dev/backend-core/internal/domain/account"
	domainOauth "github.com/Jaran-Dev/backend-core/internal/domain/oauth"
	"github.com/Jaran-Dev/backend-core/internal/infrastructure/repository"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type OAuthAccountORM struct {
	ID             uuid.UUID `gorm:"primaryKey"`
	AccountID      uuid.UUID
	Provider       string `gorm:"uniqueIndex:idx_provider_user_id,priority:1"`
	ProviderUserID string `gorm:"uniqueIndex:idx_provider_user_id,priority:2"`
	EmailSnapshot  string
	CreatedAt      time.Time
	UpdatedAt      time.Time
}

func (OAuthAccountORM) TableName() string { return "oauth_accounts" }

func (m OAuthAccountORM) toDomain() (*domainOauth.OAuthAccount, error) {
	return domainOauth.UnmarshalOAuthAccount(
		m.ID,
		domainAccount.AccountID(m.AccountID),
		domainOauth.Provider(m.Provider),
		m.ProviderUserID,
		m.EmailSnapshot,
		m.CreatedAt,
		m.UpdatedAt,
	)
}

func toPersistence(o *domainOauth.OAuthAccount) OAuthAccountORM {
	return OAuthAccountORM{
		ID:             uuid.UUID(o.ID()),
		AccountID:      uuid.UUID(o.AccountID()),
		Provider:       string(o.Provider()),
		ProviderUserID: o.ProviderUserID(),
		EmailSnapshot:  o.EmailSnapshot(),
		CreatedAt:      o.CreatedAt(),
		UpdatedAt:      o.UpdatedAt(),
	}
}

type pgRepository struct {
	db *gorm.DB
}

func NewRepository(db *gorm.DB) domainOauth.Repository {
	return &pgRepository{db: db}
}

func (r *pgRepository) Create(ctx context.Context, o *domainOauth.OAuthAccount) error {
	orm := toPersistence(o)
	db := repository.GetDB(ctx, r.db)

	if err := db.Create(&orm).Error; err != nil {
		translatedErr := repository.TranslateGormError(err)
		if errors.Is(translatedErr, gorm.ErrDuplicatedKey) {
			return domainOauth.ErrAccountAlreadyLinked
		}
		return translatedErr
	}
	return nil
}

func (r *pgRepository) GetByProvider(ctx context.Context, p domainOauth.Provider, puid string) (*domainOauth.OAuthAccount, error) {
	var orm OAuthAccountORM
	db := repository.GetDB(ctx, r.db)

	err := db.
		Where("provider = ? AND provider_user_id = ?", string(p), puid).
		First(&orm).Error
	if err != nil {
		translatedErr := repository.TranslateGormError(err)
		if errors.Is(translatedErr, gorm.ErrRecordNotFound) {
			return nil, domainOauth.ErrOAuthAccountNotFound
		}
		return nil, translatedErr
	}
	return orm.toDomain()
}

func (r *pgRepository) GetByAccountID(ctx context.Context, accID domainAccount.AccountID) ([]*domainOauth.OAuthAccount, error) {
	var orms []OAuthAccountORM
	db := repository.GetDB(ctx, r.db)

	err := db.
		Where("account_id = ?", uuid.UUID(accID)).
		Find(&orms).Error
	if err != nil {
		return nil, repository.TranslateGormError(err)
	}

	links := make([]*domainOauth.OAuthAccount, 0, len(orms))
	for _, m := range orms {
		domain, err := m.toDomain()
		if err != nil {
			return nil, err
		}
		links = append(links, domain)
	}
	return links, nil
}

func (r *pgRepository) Delete(ctx context.Context, id domainOauth.OAuthAccountID) error {
	db := repository.GetDB(ctx, r.db)

	res := db.
		Delete(&OAuthAccountORM{}, "id = ?", uuid.UUID(id))
	if res.Error != nil {
		return repository.TranslateGormError(res.Error)
	}
	if res.RowsAffected == 0 {
		return domainOauth.ErrOAuthAccountNotFound
	}
	return nil
}
