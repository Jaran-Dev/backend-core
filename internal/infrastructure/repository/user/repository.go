package user

import (
	"context"
	"errors"
	"time"

	domainAccount "github.com/Jaran-Dev/backend-core/internal/domain/account"
	domainUser "github.com/Jaran-Dev/backend-core/internal/domain/user"
	"github.com/Jaran-Dev/backend-core/internal/infrastructure/repository"
	repoAccount "github.com/Jaran-Dev/backend-core/internal/infrastructure/repository/account"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type UserORM struct {
	AccountID    uuid.UUID `gorm:"primaryKey"`
	DisplayName  string
	Username     *string    `gorm:"type:citext;uniqueIndex"`
	Bio          *string    `gorm:"type:text"`
	Headline     *string    `gorm:"type:text"`
	Pronouns     *string    `gorm:"type:varchar(20)"`
	Location     *string    `gorm:"type:varchar(100)"`
	DateOfBirth  *time.Time `gorm:"type:date"`
	AvatarURL    *string    `gorm:"type:text"`
	AvatarStatus string     `gorm:"type:varchar(20);default:pending"`
	CoverURL     *string    `gorm:"type:text"`
	CoverStatus  string     `gorm:"type:varchar(20);default:pending"`
	CreatedAt    time.Time
	UpdatedAt    time.Time
	Version      int `gorm:"default:1"`
}

func (UserORM) TableName() string { return "users" }

func (m UserORM) toDomain() (*domainUser.User, error) {
	avatarStatus := domainUser.ImageStatus(m.AvatarStatus)
	if avatarStatus == "" {
		avatarStatus = domainUser.ImageStatusPending
	}
	avatarURL := ""
	if m.AvatarURL != nil {
		avatarURL = *m.AvatarURL
	}
	avatar := domainUser.NewProfileImage(avatarURL, avatarStatus)

	coverStatus := domainUser.ImageStatus(m.CoverStatus)
	if avatarStatus == "" {
		avatarStatus = domainUser.ImageStatusPending
	}
	coverURL := ""
	if m.CoverURL != nil {
		coverURL = *m.CoverURL
	}
	cover := domainUser.NewProfileImage(coverURL, coverStatus)

	return domainUser.UnmarshalUser(
		domainAccount.AccountID(m.AccountID),
		m.DisplayName,
		m.Username,
		m.Bio,
		m.Headline,
		m.Pronouns,
		m.Location,
		m.DateOfBirth,
		avatar,
		cover,
		m.CreatedAt,
		m.UpdatedAt,
		m.Version,
	)
}

func toPersistence(u *domainUser.User) *UserORM {
	var username, bio, headline, pronouns, location *string
	var dob *time.Time
	var avatarURL, coverURL *string

	if u.Username() != "" {
		s := string(u.Username())
		username = &s
	}
	if u.Bio() != "" {
		s := string(u.Bio())
		bio = &s
	}
	if u.Headline() != "" {
		s := string(u.Headline())
		headline = &s
	}
	if u.Pronouns() != "" {
		s := string(u.Pronouns())
		pronouns = &s
	}
	if u.Location() != "" {
		s := string(u.Location())
		location = &s
	}
	if !time.Time(u.DateOfBirth()).IsZero() {
		t := u.DateOfBirth().Time()
		dob = &t
	}
	if u.Avatar().URL() != "" {
		s := u.Avatar().URL()
		avatarURL = &s
	}
	if u.Cover().URL() != "" {
		s := u.Cover().URL()
		coverURL = &s
	}

	return &UserORM{
		AccountID:    uuid.UUID(u.AccountID()),
		DisplayName:  u.DisplayName(),
		Username:     username,
		Bio:          bio,
		Headline:     headline,
		Pronouns:     pronouns,
		Location:     location,
		DateOfBirth:  dob,
		AvatarURL:    avatarURL,
		AvatarStatus: string(u.Avatar().Status()),
		CoverURL:     coverURL,
		CoverStatus:  string(u.Cover().Status()),
		CreatedAt:    u.CreatedAt(),
		UpdatedAt:    u.UpdatedAt(),
		Version:      u.Version(),
	}
}

type pgRepository struct {
	db *gorm.DB
}

func NewUserRepository(db *gorm.DB) domainUser.Repository {
	return &pgRepository{db: db}
}

func (r *pgRepository) CreateWithAccount(ctx context.Context, acc *domainAccount.Account, u *domainUser.User) error {
	accORM := repoAccount.ToPersistence(acc)
	userORM := toPersistence(u)
	db := repository.GetDB(ctx, r.db)

	err := db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(&accORM).Error; err != nil {
			translatedErr := repository.TranslateGormError(err)
			if errors.Is(translatedErr, gorm.ErrDuplicatedKey) {
				return domainAccount.ErrAccountAlreadyExists
			}
			return translatedErr
		}

		if err := tx.Create(&userORM).Error; err != nil {
			return repository.TranslateGormError(err)
		}

		return nil
	})

	return err
}

func (r *pgRepository) GetByAccountID(ctx context.Context, accID domainAccount.AccountID) (*domainUser.User, error) {
	var orm UserORM
	db := repository.GetDB(ctx, r.db)

	err := db.First(&orm, "account_id = ?", uuid.UUID(accID)).Error
	if err != nil {
		translatedErr := repository.TranslateGormError(err)
		if errors.Is(translatedErr, gorm.ErrRecordNotFound) {
			return nil, domainUser.ErrUserNotFound
		}
		return nil, translatedErr
	}

	return orm.toDomain()
}

func (r *pgRepository) GetByUsername(ctx context.Context, username string) (*domainUser.User, error) {
	var orm UserORM
	db := repository.GetDB(ctx, r.db)

	err := db.First(&orm, "username = ?", username).Error
	if err != nil {
		translatedErr := repository.TranslateGormError(err)
		if errors.Is(translatedErr, gorm.ErrRecordNotFound) {
			return nil, domainUser.ErrUserNotFound
		}
		return nil, translatedErr
	}

	return orm.toDomain()
}

func (r *pgRepository) Update(ctx context.Context, u *domainUser.User) error {
	orm := toPersistence(u)
	db := repository.GetDB(ctx, r.db)

	oldVersion := orm.Version
	orm.Version++

	result := db.Model(&UserORM{}).
		Where("account_id = ? AND version = ?", orm.AccountID, oldVersion).
		Updates(map[string]interface{}{
			"display_name":  orm.DisplayName,
			"username":      orm.Username,
			"bio":           orm.Bio,
			"headline":      orm.Headline,
			"pronouns":      orm.Pronouns,
			"location":      orm.Location,
			"date_of_birth": orm.DateOfBirth,
			"avatar_url":    orm.AvatarURL,
			"avatar_status": orm.AvatarStatus,
			"cover_url":     orm.CoverURL,
			"cover_status":  orm.CoverStatus,
			"updated_at":    orm.UpdatedAt,
			"version":       orm.Version,
		})

	if result.Error != nil {
		translatedErr := repository.TranslateGormError(result.Error)
		if errors.Is(translatedErr, gorm.ErrDuplicatedKey) {
			return domainUser.ErrUsernameTaken
		}
		return translatedErr
	}

	if result.RowsAffected == 0 {
		return domainUser.ErrUserNotFound // Or concurrent modification error if preferred
	}

	return nil
}
