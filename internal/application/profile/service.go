package profile

import (
	"context"
	"errors"
	"fmt"
	"io"
	"time"

	"github.com/Jaran-Dev/backend-core/internal/domain/account"
	domainPref "github.com/Jaran-Dev/backend-core/internal/domain/preferences"
	domainStorage "github.com/Jaran-Dev/backend-core/internal/domain/storage"
	domainUser "github.com/Jaran-Dev/backend-core/internal/domain/user"
)

type TransactionManager interface {
	Do(ctx context.Context, fn func(ctx context.Context) error) error
}

type UpdateProfileParams struct {
	DisplayName *string
	Username    *string
	Bio         *string
	Headline    *string
	Pronouns    *string
	Location    *string
	DateOfBirth *time.Time
}

type UpdatePreferencesParams struct {
	UILocale             *string
	Theme                *string
	PlaybackQuality      *string
	AutoplayNext         *bool
	SubtitleLanguage     *string
	SubtitlesEnabled     *bool
	SkipIntro            *bool
	ReducedMotion        *bool
	EmailDigestFrequency *string
	NotificationSettings *domainPref.NotificationSettings
}

type ProfileService struct {
	userRepo  domainUser.Repository
	prefRepo  domainPref.Repository
	accRepo   account.Repository
	storage   domainStorage.ObjectStorage
	txManager TransactionManager
}

func NewProfileService(
	ur domainUser.Repository,
	pr domainPref.Repository,
	ar account.Repository,
	storage domainStorage.ObjectStorage,
	txm TransactionManager,
) *ProfileService {
	return &ProfileService{
		userRepo:  ur,
		prefRepo:  pr,
		accRepo:   ar,
		storage:   storage,
		txManager: txm,
	}
}

func (s *ProfileService) GetProfile(ctx context.Context, accID account.AccountID) (*domainUser.User, *account.Account, error) {
	u, err := s.userRepo.GetByAccountID(ctx, accID)
	if err != nil {
		return nil, nil, err
	}

	a, err := s.accRepo.GetByID(ctx, accID)
	if err != nil {
		return nil, nil, err
	}

	return u, a, nil
}

func (s *ProfileService) UpdateProfile(ctx context.Context, accID account.AccountID, params UpdateProfileParams) error {
	return s.txManager.Do(ctx, func(txCtx context.Context) error {
		u, err := s.userRepo.GetByAccountID(txCtx, accID)
		if err != nil {
			return err
		}

		if params.Username != nil && *params.Username != "" {
			if string(u.Username()) != *params.Username {
				existing, err := s.userRepo.GetByUsername(txCtx, *params.Username)
				if err != nil && !errors.Is(err, domainUser.ErrUserNotFound) {
					return err
				}
				if existing != nil {
					return domainUser.ErrUsernameTaken
				}
			}
		}

		domainParams := domainUser.UpdateParams{
			DisplayName: params.DisplayName,
			Username:    params.Username,
			Bio:         params.Bio,
			Headline:    params.Headline,
			Pronouns:    params.Pronouns,
			Location:    params.Location,
			DateOfBirth: params.DateOfBirth,
		}

		if err := u.UpdateProfile(domainParams); err != nil {
			return err
		}

		return s.userRepo.Update(txCtx, u)
	})
}

func (s *ProfileService) UploadAvatar(ctx context.Context, accID account.AccountID, file io.Reader, contentType string) error {
	key := fmt.Sprintf("avatars/%s/%d", accID.String(), time.Now().Unix())

	url, err := s.storage.Put(ctx, key, file, contentType)
	if err != nil {
		return err
	}

	return s.txManager.Do(ctx, func(txCtx context.Context) error {
		u, err := s.userRepo.GetByAccountID(txCtx, accID)
		if err != nil {
			return err
		}
		u.SetAvatar(url, domainUser.ImageStatusPending)
		return s.userRepo.Update(txCtx, u)
	})
}

func (s *ProfileService) UploadCover(ctx context.Context, accID account.AccountID, file io.Reader, contentType string) error {
	key := fmt.Sprintf("covers/%s/%d", accID.String(), time.Now().Unix())

	url, err := s.storage.Put(ctx, key, file, contentType)
	if err != nil {
		return err
	}

	return s.txManager.Do(ctx, func(txCtx context.Context) error {
		u, err := s.userRepo.GetByAccountID(txCtx, accID)
		if err != nil {
			return err
		}
		u.SetCover(url, domainUser.ImageStatusPending)
		return s.userRepo.Update(txCtx, u)
	})
}

func (s *ProfileService) GetPreferences(ctx context.Context, accID account.AccountID) (*domainPref.Preferences, error) {
	p, err := s.prefRepo.GetByAccountID(ctx, accID)
	if err != nil {
		if errors.Is(err, domainPref.ErrPreferencesNotFound) {
			// Lazy migration: create default preferences if not found
			p = domainPref.NewPreferences(accID)
			if err := s.prefRepo.Create(ctx, p); err != nil {
				return nil, err
			}
			return p, nil
		}
		return nil, err
	}
	return p, nil
}

func (s *ProfileService) UpdatePreferences(ctx context.Context, accID account.AccountID, params UpdatePreferencesParams) error {
	return s.txManager.Do(ctx, func(txCtx context.Context) error {
		p, err := s.prefRepo.GetByAccountID(txCtx, accID)
		var isNew bool
		if err != nil {
			if errors.Is(err, domainPref.ErrPreferencesNotFound) {
				p = domainPref.NewPreferences(accID)
				isNew = true
			} else {
				return err
			}
		}

		domainParams := domainPref.UpdateParams{
			UILocale:             params.UILocale,
			Theme:                params.Theme,
			PlaybackQuality:      params.PlaybackQuality,
			AutoplayNext:         params.AutoplayNext,
			SubtitleLanguage:     params.SubtitleLanguage,
			SubtitlesEnabled:     params.SubtitlesEnabled,
			SkipIntro:            params.SkipIntro,
			ReducedMotion:        params.ReducedMotion,
			EmailDigestFrequency: params.EmailDigestFrequency,
			NotificationSettings: params.NotificationSettings,
		}

		if err := p.Update(domainParams); err != nil {
			return err
		}

		if isNew {
			return s.prefRepo.Create(txCtx, p)
		}
		return s.prefRepo.Update(txCtx, p)
	})
}
