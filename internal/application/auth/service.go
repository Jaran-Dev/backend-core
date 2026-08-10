package auth

import (
	"context"
	"errors"
	"time"

	"github.com/Jaran-Dev/backend-core/internal/domain/admin"
	"github.com/Jaran-Dev/backend-core/internal/domain/auth"
	"github.com/Jaran-Dev/backend-core/internal/domain/user"
	"github.com/Jaran-Dev/backend-core/internal/infrastructure/cache"

	"github.com/google/uuid"
)

type AuthService struct {
	userRepo  user.UserRepository
	adminRepo admin.AdminRepository
	hasher    auth.Hasher
	tokenSvc  auth.TokenService
	store     cache.RedisSessionStore

	refreshTTL time.Duration
}

func NewAuthService(ur user.UserRepository, ar admin.AdminRepository, h auth.Hasher, ts auth.TokenService, store cache.RedisSessionStore, refreshTTL time.Duration) *AuthService {
	return &AuthService{
		userRepo:   ur,
		adminRepo:  ar,
		hasher:     h,
		tokenSvc:   ts,
		store:      store,
		refreshTTL: refreshTTL,
	}
}

func (s *AuthService) Register(ctx context.Context, email, password, displayName, uiLocale string) (*user.User, error) {
	e, err := user.NewEmail(email)
	if err != nil {
		return nil, err
	}

	_, err = s.userRepo.GetByEmail(ctx, e)
	if err == nil {
		return nil, user.ErrUserAlreadyExists
	} else if !errors.Is(err, user.ErrUserNotFound) {
		return nil, err
	}

	hash, err := s.hasher.Hash(password)
	if err != nil {
		return nil, err
	}

	u := user.NewUser(e, hash, displayName, uiLocale)
	if err := s.userRepo.Create(ctx, u); err != nil {
		return nil, err
	}
	return u, nil
}

func (s *AuthService) LoginUser(ctx context.Context, email, password string) (string, string, time.Time, error) {
	e, err := user.NewEmail(email)
	if err != nil {
		return "", "", time.Time{}, err
	}

	u, err := s.userRepo.GetByEmail(ctx, e)
	if err != nil {
		return "", "", time.Time{}, auth.ErrInvalidCredentials
	}

	match, err := s.hasher.Compare(u.PasswordHash(), password)
	if err != nil || !match {
		return "", "", time.Time{}, auth.ErrInvalidCredentials
	}

	// TODO
	// return s.issueTokens(ctx, u.ID(), auth.RoleUser)

	// !!! PLACEHOLDER
	return s.issueTokens(ctx, uuid.New(), auth.RoleUser)
}

func (s *AuthService) LoginAdmin(ctx context.Context, email, password string) (string, string, time.Time, error) {
	e, err := admin.NewEmail(email)
	if err != nil {
		return "", "", time.Time{}, err
	}

	a, err := s.adminRepo.GetByEmail(ctx, e)
	if err != nil {
		return "", "", time.Time{}, auth.ErrInvalidCredentials
	}

	match, err := s.hasher.Compare(a.PasswordHash(), password)
	if err != nil || !match {
		return "", "", time.Time{}, auth.ErrInvalidCredentials
	}

	// TODO
	// return s.issueTokens(ctx, a.ID(), auth.RoleAdmin)

	// !!! PLACEHOLDER
	return s.issueTokens(ctx, uuid.New(), auth.RoleAdmin)
}

func (s *AuthService) Refresh(ctx context.Context, rawRefreshToken string) (string, string, time.Time, error) {
	hashedOld := cache.HashToken(rawRefreshToken)

	// Generate new raw token upfront
	newRawToken, err := cache.GenerateOpaqueToken()
	if err != nil {
		return "", "", time.Time{}, err
	}
	hashedNew := cache.HashToken(newRawToken)

	newExp := time.Now().Add(s.refreshTTL)
	uid, role, err := s.store.AtomicRotate(ctx, hashedOld, hashedNew, newExp)
	if err != nil {
		return "", "", time.Time{}, err
	}

	accessToken, _, err := s.tokenSvc.GenerateAccessToken(uid, role)
	if err != nil {
		return "", "", time.Time{}, err
	}

	return accessToken, newRawToken, newExp, nil
}

func (s *AuthService) Logout(ctx context.Context, rawRefreshToken string) error {
	hashed := cache.HashToken(rawRefreshToken)
	_, _, familyID, _, _, err := s.store.GetSession(ctx, hashed)
	if err != nil {
		return err // silently fail or log
	}
	return s.store.RevokeFamily(ctx, familyID)
}

func (s *AuthService) issueTokens(ctx context.Context, userID uuid.UUID, role auth.Role) (string, string, time.Time, error) {
	accessToken, _, err := s.tokenSvc.GenerateAccessToken(userID, role)
	if err != nil {
		return "", "", time.Time{}, err
	}

	rawRefresh, err := cache.GenerateOpaqueToken()
	if err != nil {
		return "", "", time.Time{}, err
	}
	hashedRefresh := cache.HashToken(rawRefresh)
	familyID := uuid.NewString()
	expiresAt := time.Now().Add(s.refreshTTL)

	if err := s.store.SaveSession(ctx, hashedRefresh, userID, role, familyID, expiresAt); err != nil {
		return "", "", time.Time{}, err
	}

	return accessToken, rawRefresh, expiresAt, nil
}
