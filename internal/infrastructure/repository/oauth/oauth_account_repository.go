package oauth

import (
	"context"
	"errors"
	"time"

	domainAccount "github.com/Jaran-Dev/backend-core/internal/domain/account"
	domainOauth "github.com/Jaran-Dev/backend-core/internal/domain/oauth"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type OAuthAccountModel struct {
	ID             uuid.UUID `gorm:"primaryKey"`
	AccountID      uuid.UUID
	Provider       string `gorm:"uniqueIndex:idx_provider_user_id,priority:1"`
	ProviderUserID string `gorm:"uniqueIndex:idx_provider_user_id,priority:2"`
	EmailSnapshot  string
	CreatedAt      time.Time
	UpdatedAt      time.Time
}

func (OAuthAccountModel) TableName() string { return "oauth_accounts" }

func (m OAuthAccountModel) toDomain() (*domainOauth.OAuthAccount, error) {
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

func fromDomain(o *domainOauth.OAuthAccount) OAuthAccountModel {
	return OAuthAccountModel{
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
	model := fromDomain(o)
	if err := r.db.WithContext(ctx).Create(&model).Error; err != nil {
		if errors.Is(err, gorm.ErrDuplicatedKey) {
			return domainOauth.ErrAccountAlreadyLinked
		}
		return err
	}
	return nil
}

func (r *pgRepository) GetByProvider(ctx context.Context, p domainOauth.Provider, puid string) (*domainOauth.OAuthAccount, error) {
	var model OAuthAccountModel
	err := r.db.WithContext(ctx).
		Where("provider = ? AND provider_user_id = ?", string(p), puid).
		First(&model).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, domainOauth.ErrOAuthAccountNotFound
		}
		return nil, err
	}
	return model.toDomain()
}

func (r *pgRepository) GetByAccountID(ctx context.Context, accID domainAccount.AccountID) ([]*domainOauth.OAuthAccount, error) {
	var models []OAuthAccountModel
	err := r.db.WithContext(ctx).
		Where("account_id = ?", uuid.UUID(accID)).
		Find(&models).Error
	if err != nil {
		return nil, err
	}

	links := make([]*domainOauth.OAuthAccount, 0, len(models))
	for _, m := range models {
		domain, err := m.toDomain()
		if err != nil {
			return nil, err
		}
		links = append(links, domain)
	}
	return links, nil
}

func (r *pgRepository) Delete(ctx context.Context, id domainOauth.OAuthAccountID) error {
	res := r.db.WithContext(ctx).
		Delete(&OAuthAccountModel{}, "id = ?", uuid.UUID(id))
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return domainOauth.ErrOAuthAccountNotFound
	}
	return nil
}
