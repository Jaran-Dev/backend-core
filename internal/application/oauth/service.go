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
	tokenUtil "github.com/Jaran-Dev/backend-core/internal/pkg/token"

	"github.com/google/uuid"
)

type TransactionManager interface {
	Do(ctx context.Context, fn func(ctx context.Context) error) error
}

type CallbackResult struct {
	AccessToken  string
	RefreshToken string
	RedirectURL  string
	ExpiresAt    time.Time
}

type OAuthService struct {
	providers    map[domainOauth.Provider]domainOauth.OAuthProvider
	stateStore   domainOauth.StateStore
	oauthRepo    domainOauth.Repository
	accountRepo  domainAccount.Repository
	userRepo     domainUser.Repository
	txManager    TransactionManager
	tokenSvc     domainAuth.TokenService
	sessionStore domainAuth.SessionStore
	refreshTTL   time.Duration
}

func NewOAuthService(
	providers map[domainOauth.Provider]domainOauth.OAuthProvider,
	stateStore domainOauth.StateStore,
	oauthRepo domainOauth.Repository,
	accountRepo domainAccount.Repository,
	userRepo domainUser.Repository,
	txm TransactionManager,
	tokenSvc domainAuth.TokenService,
	sessionStore domainAuth.SessionStore,
	refreshTTL time.Duration,
) *OAuthService {
	return &OAuthService{
		providers:    providers,
		stateStore:   stateStore,
		oauthRepo:    oauthRepo,
		accountRepo:  accountRepo,
		userRepo:     userRepo,
		txManager:    txm,
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

func (s *OAuthService) HandleCallback(ctx context.Context, providerName domainOauth.Provider, code, state string, meta domainAuth.SessionMetadata) (*CallbackResult, error) {
	provider, ok := s.providers[providerName]
	if !ok {
		return nil, domainOauth.ErrOAuthProviderUnsupported
	}

	codeVerifier, redirectURL, err := s.stateStore.Consume(ctx, state)
	if err != nil {
		return nil, domainOauth.ErrInvalidState
	}

	profile, err := provider.Exchange(ctx, code, codeVerifier)
	if err != nil {
		return nil, domainOauth.ErrOAuthCodeExchangeFailed
	}

	if profile == nil {
		return nil, domainOauth.ErrOAuthCodeExchangeFailed
	}

	accID, err := s.resolveOAuthAccountID(ctx, providerName, *profile)
	if err != nil {
		return nil, err
	}

	acc, err := s.accountRepo.GetByID(ctx, accID)
	if err != nil {
		return nil, err
	}

	switch acc.Status() {
	case domainAccount.StatusBanned:
		return nil, domainAccount.ErrAccountBanned
	case domainAccount.StatusLocked:
		return nil, domainAccount.ErrAccountLocked
	case domainAccount.StatusPendingVerification:
		acc.Verify()
		if err := s.accountRepo.Update(ctx, acc); err != nil {
			return nil, err
		}
	}

	access, refresh, exp, err := s.issueTokens(ctx, acc.ID(), domainAuth.Role(acc.Role()), meta)
	if err != nil {
		return nil, err
	}

	return &CallbackResult{
		AccessToken:  access,
		RefreshToken: refresh,
		RedirectURL:  redirectURL,
		ExpiresAt:    exp,
	}, nil
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

	if profile == nil {
		return domainOauth.ErrOAuthCodeExchangeFailed
	}

	return s.linkOAuthAccount(ctx, accID, providerName, *profile)
}

func (s *OAuthService) UnlinkAccount(ctx context.Context, accID domainAccount.AccountID, providerName domainOauth.Provider) error {
	return s.txManager.Do(ctx, func(txCtx context.Context) error {
		links, err := s.oauthRepo.GetByAccountID(txCtx, accID)
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

		acc, err := s.accountRepo.GetByID(txCtx, accID)
		if err != nil {
			return err
		}

		// Invariant check: Must have a password OR more than 1 OAuth link
		if acc.PasswordHash() == nil && len(links) == 1 {
			return domainOauth.ErrCannotUnlinkLastMethod
		}

		return s.oauthRepo.Delete(txCtx, toDelete.ID())
	})
}

func (s *OAuthService) resolveOAuthAccountID(ctx context.Context, providerName domainOauth.Provider, profile domainOauth.Profile) (domainAccount.AccountID, error) {
	var accID domainAccount.AccountID

	err := s.txManager.Do(ctx, func(txCtx context.Context) error {
		existingOauth, err := s.oauthRepo.GetByProvider(txCtx, providerName, profile.ProviderUserID)
		if err == nil {
			accID = existingOauth.AccountID()
			return nil
		}
		if !errors.Is(err, domainOauth.ErrOAuthAccountNotFound) {
			return err
		}

		existingAcc, err := s.accountRepo.GetByEmail(txCtx, profile.Email)
		if err != nil && !errors.Is(err, domainAccount.ErrAccountNotFound) {
			return err
		}

		if err == nil {
			// User exists by email, just link OAuth
			accID = existingAcc.ID()
			return s.linkOAuthAccount(txCtx, accID, providerName, profile)
		}

		// Create new account and user
		acc, err := domainAccount.NewAccount(profile.Email, nil, domainAccount.RoleUser)
		if err != nil {
			return err
		}
		acc.Verify()
		u, err := domainUser.NewUser(acc.ID(), profile.DisplayName)
		if err != nil {
			return err
		}

		if err := s.userRepo.CreateWithAccount(txCtx, acc, u); err != nil {
			return err
		}

		accID = acc.ID()
		return s.linkOAuthAccount(txCtx, accID, providerName, profile)
	})

	if err != nil {
		return domainAccount.AccountID(uuid.Nil), err
	}
	return accID, nil
}

func (s *OAuthService) linkOAuthAccount(ctx context.Context, accID domainAccount.AccountID, providerName domainOauth.Provider, profile domainOauth.Profile) error {
	oauthAcc := domainOauth.NewOAuthAccount(accID, providerName, profile.ProviderUserID, profile.Email)

	err := s.oauthRepo.Create(ctx, oauthAcc)
	if err == nil {
		return nil
	}

	if errors.Is(err, domainOauth.ErrAccountAlreadyLinked) {
		existing, fetchErr := s.oauthRepo.GetByProvider(ctx, providerName, profile.ProviderUserID)
		if fetchErr != nil {
			return fetchErr
		}

		if existing.AccountID() != accID {
			return domainOauth.ErrAccountAlreadyLinked
		}

		return nil
	}

	return err
}

func (s *OAuthService) issueTokens(ctx context.Context, accountID domainAccount.AccountID, role domainAuth.Role, meta domainAuth.SessionMetadata) (string, string, time.Time, error) {
	accessToken, _, err := s.tokenSvc.GenerateAccessToken(accountID, role)
	if err != nil {
		return "", "", time.Time{}, err
	}

	rawRefresh, err := tokenUtil.GenerateOpaqueToken()
	if err != nil {
		return "", "", time.Time{}, err
	}
	hashedRefresh := tokenUtil.HashToken(rawRefresh)

	sessionID := uuid.New()
	expiresAt := time.Now().Add(s.refreshTTL)

	if err := s.sessionStore.SaveSession(ctx, sessionID, hashedRefresh, accountID, role, expiresAt, meta); err != nil {
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
