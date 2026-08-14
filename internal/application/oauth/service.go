package oauth

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"errors"
	"time"

	domainAccount "github.com/Jaran-Dev/backend-core/internal/domain/account"
	domainAuth "github.com/Jaran-Dev/backend-core/internal/domain/auth"
	domainOauth "github.com/Jaran-Dev/backend-core/internal/domain/oauth"
	domainUser "github.com/Jaran-Dev/backend-core/internal/domain/user"
	"github.com/Jaran-Dev/backend-core/internal/infrastructure/cache"

	"github.com/google/uuid"
)

type OAuthService struct {
	providers    map[domainOauth.Provider]domainOauth.OAuthProvider
	stateStore   domainOauth.StateStore
	oauthRepo    domainOauth.Repository
	accountRepo  domainAccount.Repository
	userRepo     domainUser.Repository
	tokenSvc     domainAuth.TokenService
	sessionStore cache.RedisSessionStore
	refreshTTL   time.Duration
}

func NewOAuthService(
	providers map[domainOauth.Provider]domainOauth.OAuthProvider,
	stateStore domainOauth.StateStore,
	oauthRepo domainOauth.Repository,
	accountRepo domainAccount.Repository,
	userRepo domainUser.Repository,
	tokenSvc domainAuth.TokenService,
	sessionStore cache.RedisSessionStore,
	refreshTTL time.Duration,
) *OAuthService {
	return &OAuthService{
		providers:    providers,
		stateStore:   stateStore,
		oauthRepo:    oauthRepo,
		accountRepo:  accountRepo,
		userRepo:     userRepo,
		tokenSvc:     tokenSvc,
		sessionStore: sessionStore,
		refreshTTL:   refreshTTL,
	}
}

func (s *OAuthService) InitiateLogin(ctx context.Context, providerName domainOauth.Provider, redirectURL string) (string, error) {
	provider, ok := s.providers[providerName]
	if !ok {
		return "", domainOauth.ErrOAuthProviderUnsupported
	}

	state := generateRandomString(32)
	codeVerifier := generateRandomString(64)
	codeChallenge := generateCodeChallenge(codeVerifier)

	if err := s.stateStore.Save(ctx, state, codeVerifier, redirectURL); err != nil {
		return "", err
	}

	return provider.AuthURL(state, codeChallenge), nil
}

func (s *OAuthService) HandleCallback(ctx context.Context, providerName domainOauth.Provider, code, state string) (string, string, time.Time, error) {
	provider, ok := s.providers[providerName]
	if !ok {
		return "", "", time.Time{}, domainOauth.ErrOAuthProviderUnsupported
	}

	codeVerifier, _, err := s.stateStore.Consume(ctx, state)
	if err != nil {
		return "", "", time.Time{}, domainOauth.ErrInvalidState
	}

	profile, err := provider.Exchange(ctx, code, codeVerifier)
	if err != nil {
		return "", "", time.Time{}, domainOauth.ErrOAuthCodeExchangeFailed
	}

	var accID domainAccount.AccountID

	existingOauth, err := s.oauthRepo.GetByProvider(ctx, providerName, profile.ProviderUserID)
	if err == nil {
		accID = existingOauth.AccountID()
	} else if errors.Is(err, domainOauth.ErrOAuthAccountNotFound) {
		existingAcc, err := s.accountRepo.GetByEmail(ctx, profile.Email)
		if err != nil && !errors.Is(err, domainAccount.ErrAccountNotFound) {
			return "", "", time.Time{}, err
		}

		if err == nil {
			accID = existingAcc.ID()
		} else {
			acc, err := domainAccount.NewAccount(profile.Email, nil, domainAccount.RoleUser)
			if err != nil {
				return "", "", time.Time{}, err
			}
			acc.Verify()
			accID = acc.ID()

			u := domainUser.NewUser(acc.ID(), profile.DisplayName, profile.Locale)
			if err := s.userRepo.CreateWithAccount(ctx, acc, u); err != nil {
				return "", "", time.Time{}, err
			}
		}

		oauthAcc := domainOauth.NewOAuthAccount(accID, providerName, profile.ProviderUserID, profile.Email)
		if err := s.oauthRepo.Create(ctx, oauthAcc); err != nil {
			if errors.Is(err, domainOauth.ErrAccountAlreadyLinked) {
				existing, fetchErr := s.oauthRepo.GetByProvider(ctx, providerName, profile.ProviderUserID)
				if fetchErr != nil {
					return "", "", time.Time{}, fetchErr
				}
				accID = existing.AccountID()
			} else {
				return "", "", time.Time{}, err
			}
		}
	} else {
		return "", "", time.Time{}, err
	}

	return s.issueTokens(ctx, accID, domainAuth.RoleUser)
}

func (s *OAuthService) LinkAccount(ctx context.Context, accID domainAccount.AccountID, providerName domainOauth.Provider, code, state string) error {
	provider, ok := s.providers[providerName]
	if !ok {
		return domainOauth.ErrOAuthProviderUnsupported
	}

	codeVerifier, _, err := s.stateStore.Consume(ctx, state)
	if err != nil {
		return domainOauth.ErrInvalidState
	}

	profile, err := provider.Exchange(ctx, code, codeVerifier)
	if err != nil {
		return domainOauth.ErrOAuthCodeExchangeFailed
	}

	existing, err := s.oauthRepo.GetByProvider(ctx, providerName, profile.ProviderUserID)
	if err == nil && existing.AccountID() != accID {
		return domainOauth.ErrAccountAlreadyLinked
	}
	if err != nil && !errors.Is(err, domainOauth.ErrOAuthAccountNotFound) {
		return err
	}

	oauthAcc := domainOauth.NewOAuthAccount(accID, providerName, profile.ProviderUserID, profile.Email)
	return s.oauthRepo.Create(ctx, oauthAcc)
}

func (s *OAuthService) UnlinkAccount(ctx context.Context, accID domainAccount.AccountID, providerName domainOauth.Provider) error {
	links, err := s.oauthRepo.GetByAccountID(ctx, accID)
	if err != nil {
		return err
	}

	var toDelete *domainOauth.OAuthAccount
	for _, l := range links {
		if l.Provider() == providerName {
			toDelete = l
			break
		}
	}

	if toDelete == nil {
		return domainOauth.ErrOAuthAccountNotFound
	}

	acc, err := s.accountRepo.GetByID(ctx, accID)
	if err != nil {
		return err
	}

	// Domain rule: cannot unlink if no password and this is the last OAuth link
	if acc.PasswordHash() == nil && len(links) == 1 {
		return domainOauth.ErrCannotUnlinkLastMethod
	}

	return s.oauthRepo.Delete(ctx, toDelete.ID())
}

func (s *OAuthService) issueTokens(ctx context.Context, accountID domainAccount.AccountID, role domainAuth.Role) (string, string, time.Time, error) {
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

	if err := s.sessionStore.SaveSession(ctx, hashedRefresh, accountID, role, familyID, expiresAt); err != nil {
		return "", "", time.Time{}, err
	}

	return accessToken, rawRefresh, expiresAt, nil
}

func generateRandomString(length int) string {
	b := make([]byte, length)
	_, _ = rand.Read(b)
	return base64.RawURLEncoding.EncodeToString(b)
}

func generateCodeChallenge(verifier string) string {
	h := sha256.Sum256([]byte(verifier))
	return base64.RawURLEncoding.EncodeToString(h[:])
}
