package auth

import (
	"context"
	"errors"
	"time"

	"github.com/Jaran-Dev/backend-core/internal/domain/account"
	"github.com/Jaran-Dev/backend-core/internal/domain/auth"
	"github.com/Jaran-Dev/backend-core/internal/domain/user"
	"github.com/Jaran-Dev/backend-core/internal/infrastructure/cache"

	"github.com/google/uuid"
)

type AuthService struct {
	accountRepo account.Repository
	userRepo    user.Repository
	hasher      auth.Hasher
	tokenSvc    auth.TokenService
	store       cache.RedisSessionStore
	refreshTTL  time.Duration
}

func NewAuthService(ar account.Repository, ur user.Repository, h auth.Hasher, ts auth.TokenService, store cache.RedisSessionStore, refreshTTL time.Duration) *AuthService {
	return &AuthService{
		accountRepo: ar,
		userRepo:    ur,
		hasher:      h,
		tokenSvc:    ts,
		store:       store,
		refreshTTL:  refreshTTL,
	}
}

func (s *AuthService) Register(ctx context.Context, email, password, displayName, uiLocale string) (*user.User, error) {
	_, err := s.accountRepo.GetByEmail(ctx, email)
	if err == nil {
		return nil, account.ErrAccountAlreadyExists
	} else if !errors.Is(err, account.ErrAccountNotFound) {
		return nil, err
	}

	hash, err := s.hasher.Hash(password)
	if err != nil {
		return nil, err
	}

	acc, err := account.NewAccount(email, hash, account.RoleUser)
	if err != nil {
		return nil, err
	}

	u := user.NewUser(acc.ID(), displayName, uiLocale)

	if err := s.userRepo.CreateWithAccount(ctx, acc, u); err != nil {
		return nil, err
	}

	return u, nil
}

func (s *AuthService) Login(ctx context.Context, email, password string) (string, string, time.Time, error) {
	acc, err := s.accountRepo.GetByEmail(ctx, email)
	if err != nil {
		return "", "", time.Time{}, auth.ErrInvalidCredentials
	}

	if acc.Status() == account.StatusBanned {
		return "", "", time.Time{}, account.ErrAccountBanned
	}

	match, err := s.hasher.Compare(acc.PasswordHash(), password)
	if err != nil || !match {
		return "", "", time.Time{}, auth.ErrInvalidCredentials
	}

	authRole := auth.Role(acc.Role())
	return s.issueTokens(ctx, acc.ID(), authRole)
}

func (s *AuthService) Refresh(ctx context.Context, rawRefreshToken string) (string, string, time.Time, error) {
	hashedOld := cache.HashToken(rawRefreshToken)

	newRawToken, err := cache.GenerateOpaqueToken()
	if err != nil {
		return "", "", time.Time{}, err
	}
	hashedNew := cache.HashToken(newRawToken)

	newExp := time.Now().Add(s.refreshTTL)
	accID, role, err := s.store.AtomicRotate(ctx, hashedOld, hashedNew, newExp)
	if err != nil {
		return "", "", time.Time{}, err
	}

	accessToken, _, err := s.tokenSvc.GenerateAccessToken(accID, role)
	if err != nil {
		return "", "", time.Time{}, err
	}

	return accessToken, newRawToken, newExp, nil
}

func (s *AuthService) Logout(ctx context.Context, rawRefreshToken string) error {
	hashed := cache.HashToken(rawRefreshToken)
	_, _, familyID, _, _, err := s.store.GetSession(ctx, hashed)
	if err != nil {
		return err
	}
	return s.store.RevokeFamily(ctx, familyID)
}

func (s *AuthService) issueTokens(ctx context.Context, accountID account.AccountID, role auth.Role) (string, string, time.Time, error) {
	accessToken, _, err := s.tokenSvc.GenerateAccessToken(accountID, role)
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

	if err := s.store.SaveSession(ctx, hashedRefresh, accountID, role, familyID, expiresAt); err != nil {
		return "", "", time.Time{}, err
	}

	return accessToken, rawRefresh, expiresAt, nil
}
