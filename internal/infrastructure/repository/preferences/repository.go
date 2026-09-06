package preferences

import (
	"context"
	"encoding/json"
	"errors"
	"time"

	domainAccount "github.com/Jaran-Dev/backend-core/internal/domain/account"
	domainPref "github.com/Jaran-Dev/backend-core/internal/domain/preferences"
	"github.com/Jaran-Dev/backend-core/internal/infrastructure/repository"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type PreferenceORM struct {
	AccountID            uuid.UUID `gorm:"primaryKey"`
	UILocale             string    `gorm:"type:varchar(5)"`
	Theme                string    `gorm:"type:varchar(10)"`
	PlaybackQuality      string    `gorm:"type:varchar(10)"`
	AutoplayNext         bool
	SubtitleLanguage     string `gorm:"type:varchar(5)"`
	SubtitlesEnabled     bool
	SkipIntro            bool
	ReducedMotion        bool
	EmailDigestFrequency string          `gorm:"type:varchar(10)"`
	NotificationSettings json.RawMessage `gorm:"type:jsonb"`
	CreatedAt            time.Time
	UpdatedAt            time.Time
	Version              int `gorm:"default:1"`
}

func (PreferenceORM) TableName() string { return "preferences" }

func (m PreferenceORM) toDomain() (*domainPref.Preferences, error) {
	var notif domainPref.NotificationSettings
	if len(m.NotificationSettings) > 0 {
		if err := json.Unmarshal(m.NotificationSettings, &notif); err != nil {
			return nil, err
		}
	}

	return domainPref.UnmarshalPreferences(
		domainAccount.AccountID(m.AccountID),
		m.UILocale,
		m.Theme,
		m.PlaybackQuality,
		m.AutoplayNext,
		m.SubtitleLanguage,
		m.SubtitlesEnabled,
		m.SkipIntro,
		m.ReducedMotion,
		m.EmailDigestFrequency,
		notif,
		m.CreatedAt,
		m.UpdatedAt,
		m.Version,
	)
}

func toPersistence(p *domainPref.Preferences) *PreferenceORM {
	notifJSON, _ := json.Marshal(p.NotificationSettings())
	return &PreferenceORM{
		AccountID:            uuid.UUID(p.AccountID()),
		UILocale:             string(p.UILocale()),
		Theme:                string(p.Theme()),
		PlaybackQuality:      string(p.PlaybackQuality()),
		AutoplayNext:         p.AutoplayNext(),
		SubtitleLanguage:     string(p.SubtitleLanguage()),
		SubtitlesEnabled:     p.SubtitlesEnabled(),
		SkipIntro:            p.SkipIntro(),
		ReducedMotion:        p.ReducedMotion(),
		EmailDigestFrequency: string(p.EmailDigestFrequency()),
		NotificationSettings: notifJSON,
		CreatedAt:            p.CreatedAt(),
		UpdatedAt:            p.UpdatedAt(),
		Version:              p.Version(),
	}
}

type pgRepository struct {
	db *gorm.DB
}

func NewPreferencesRepository(db *gorm.DB) domainPref.Repository {
	return &pgRepository{db: db}
}

func (r *pgRepository) Create(ctx context.Context, p *domainPref.Preferences) error {
	orm := toPersistence(p)
	db := repository.GetDB(ctx, r.db)

	if err := db.Create(&orm).Error; err != nil {
		return repository.TranslateGormError(err)
	}
	return nil
}

func (r *pgRepository) GetByAccountID(ctx context.Context, accID domainAccount.AccountID) (*domainPref.Preferences, error) {
	var orm PreferenceORM
	db := repository.GetDB(ctx, r.db)

	err := db.First(&orm, "account_id = ?", uuid.UUID(accID)).Error
	if err != nil {
		translatedErr := repository.TranslateGormError(err)
		if errors.Is(translatedErr, gorm.ErrRecordNotFound) {
			return nil, domainPref.ErrPreferencesNotFound
		}
		return nil, translatedErr
	}

	return orm.toDomain()
}

func (r *pgRepository) Update(ctx context.Context, p *domainPref.Preferences) error {
	orm := toPersistence(p)
	db := repository.GetDB(ctx, r.db)

	oldVersion := orm.Version
	orm.Version++

	result := db.Model(&PreferenceORM{}).
		Where("account_id = ? AND version = ?", orm.AccountID, oldVersion).
		Updates(map[string]interface{}{
			"ui_locale":              orm.UILocale,
			"theme":                  orm.Theme,
			"playback_quality":       orm.PlaybackQuality,
			"autoplay_next":          orm.AutoplayNext,
			"subtitle_language":      orm.SubtitleLanguage,
			"subtitles_enabled":      orm.SubtitlesEnabled,
			"skip_intro":             orm.SkipIntro,
			"reduced_motion":         orm.ReducedMotion,
			"email_digest_frequency": orm.EmailDigestFrequency,
			"notification_settings":  orm.NotificationSettings,
			"updated_at":             orm.UpdatedAt,
			"version":                orm.Version,
		})

	if result.Error != nil {
		return repository.TranslateGormError(result.Error)
	}

	if result.RowsAffected == 0 {
		return domainPref.ErrConcurrentModification
	}

	return nil
}
