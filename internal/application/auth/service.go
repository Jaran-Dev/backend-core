package auth

import (
	"context"
	"errors"
	"time"

	"github.com/Jaran-Dev/backend-core/internal/domain/account"
	"github.com/Jaran-Dev/backend-core/internal/domain/auth"
	"github.com/Jaran-Dev/backend-core/internal/domain/user"
	tokenUtil "github.com/Jaran-Dev/backend-core/internal/pkg/token"
	"github.com/google/uuid"
)

type AuthService struct {
	accountRepo   account.Repository
	userRepo      user.Repository
	hasher        auth.Hasher
	tokenSvc      auth.TokenService
	store         auth.SessionStore
	emailSvc      auth.EmailService
	attemptStore  auth.LoginAttemptStore
	verifyTTL     time.Duration
	refreshTTL    time.Duration
	attemptWindow time.Duration
	maxAttempts   int
}

func NewAuthService(
	ar account.Repository,
	ur user.Repository,
	h auth.Hasher,
	ts auth.TokenService,
	store auth.SessionStore,
	emailSvc auth.EmailService,
	attemptStore auth.LoginAttemptStore,
	verifyTTL time.Duration,
	refreshTTL time.Duration,
	attemptWindow time.Duration,
	maxAttempts int,
) *AuthService {
	return &AuthService{
		accountRepo:   ar,
		userRepo:      ur,
		hasher:        h,
		tokenSvc:      ts,
		store:         store,
		emailSvc:      emailSvc,
		attemptStore:  attemptStore,
		verifyTTL:     verifyTTL,
		refreshTTL:    refreshTTL,
		attemptWindow: attemptWindow,
		maxAttempts:   maxAttempts,
	}
}

func (s *AuthService) Register(ctx context.Context, email, password, displayName string) (*user.User, error) {
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

	acc, err := account.NewAccount(email, &hash, account.RoleUser)
	if err != nil {
		return nil, err
	}

	u, err := user.NewUser(acc.ID(), displayName)
	if err != nil {
		return nil, err
	}

	if err := s.userRepo.CreateWithAccount(ctx, acc, u); err != nil {
		return nil, err
	}

	verificationToken, err := s.tokenSvc.GenerateEmailVerificationToken(acc.ID())
	if err != nil {
		return nil, err
	}

	if err := s.emailSvc.SendVerificationEmail(email, verificationToken); err != nil {
		// Log error
	}

	return u, nil
}

func (s *AuthService) Login(ctx context.Context, email, password string, meta auth.SessionMetadata) (string, string, time.Time, error) {
	acc, err := s.accountRepo.GetByEmail(ctx, email)
	if err != nil {
		return "", "", time.Time{}, auth.ErrInvalidCredentials
	}

	if acc.PasswordHash() == nil {
		return "", "", time.Time{}, auth.ErrInvalidCredentials
	}

	if acc.Status() == account.StatusLocked {
		isLocked, err := s.attemptStore.IsAccountLocked(ctx, acc.ID())
		if err != nil {
			return "", "", time.Time{}, err
		}
		if isLocked {
			return "", "", time.Time{}, account.ErrAccountLocked
		}
		acc.Unlock()
		if err := s.accountRepo.Update(ctx, acc); err != nil {
			return "", "", time.Time{}, err
		}
	}

	if acc.Status() == account.StatusBanned {
		return "", "", time.Time{}, account.ErrAccountBanned
	}

	match, err := s.hasher.Compare(*acc.PasswordHash(), password)
	if err != nil || !match {
		attempts, _ := s.attemptStore.IncrementFailedAttempt(ctx, email)
		if attempts >= s.maxAttempts {
			acc.Lock()
			_ = s.accountRepo.Update(ctx, acc)
			_ = s.attemptStore.LockAccount(ctx, acc.ID(), s.attemptWindow)
			return "", "", time.Time{}, account.ErrAccountLocked
		}
		return "", "", time.Time{}, auth.ErrInvalidCredentials
	}

	_ = s.attemptStore.ResetFailedAttempts(ctx, email)

	authRole := auth.Role(acc.Role())
	return s.issueTokens(ctx, acc.ID(), authRole, meta)
}

func (s *AuthService) Refresh(ctx context.Context, rawRefreshToken string, meta auth.SessionMetadata) (string, string, time.Time, error) {
	hashedOld := tokenUtil.HashToken(rawRefreshToken)

	newRawToken, err := tokenUtil.GenerateOpaqueToken()
	if err != nil {
		return "", "", time.Time{}, err
	}
	hashedNew := tokenUtil.HashToken(newRawToken)

	newExp := time.Now().Add(s.refreshTTL)
	_, accID, role, err := s.store.AtomicRotate(ctx, hashedOld, hashedNew, newExp, meta)
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
	hashed := tokenUtil.HashToken(rawRefreshToken)

	sessionID, accID, _, _, _, err := s.store.GetSession(ctx, hashed)
	if err != nil {
		return err
	}

	return s.store.RevokeSessionByID(ctx, accID, sessionID)
}

func (s *AuthService) ListSessions(ctx context.Context, accID account.AccountID) ([]*auth.Session, error) {
	return s.store.ListSessionsByAccountID(ctx, accID)
}

func (s *AuthService) RevokeSession(ctx context.Context, accID account.AccountID, sessionID uuid.UUID) error {
	return s.store.RevokeSessionByID(ctx, accID, sessionID)
}

func (s *AuthService) issueTokens(ctx context.Context, accountID account.AccountID, role auth.Role, meta auth.SessionMetadata) (string, string, time.Time, error) {
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

	if err := s.store.SaveSession(ctx, sessionID, hashedRefresh, accountID, role, expiresAt, meta); err != nil {
		return "", "", time.Time{}, err
	}

	return accessToken, rawRefresh, expiresAt, nil
}

func (s *AuthService) VerifyEmail(ctx context.Context, token string) error {
	accID, err := s.tokenSvc.ValidateEmailVerificationToken(token)
	if err != nil {
		return auth.ErrInvalidCredentials
	}

	acc, err := s.accountRepo.GetByID(ctx, accID)
	if err != nil {
		return err
	}

	if acc.Status() == account.StatusActive {
		return nil
	}

	acc.Verify()
	return s.accountRepo.Update(ctx, acc)
}

func (s *AuthService) ResendVerification(ctx context.Context, email string) error {
	acc, err := s.accountRepo.GetByEmail(ctx, email)
	if err != nil {
		return nil
	}

	if acc.Status() == account.StatusActive {
		return nil
	}

	verificationToken, err := s.tokenSvc.GenerateEmailVerificationToken(acc.ID())
	if err != nil {
		return err
	}

	return s.emailSvc.SendVerificationEmail(email, verificationToken)
}

func (s *AuthService) ForgotPassword(ctx context.Context, email string) error {
	acc, err := s.accountRepo.GetByEmail(ctx, email)
	if err != nil {
		return nil
	}

	rawToken, err := tokenUtil.GenerateOpaqueToken()
	if err != nil {
		return err
	}
	hashedToken := tokenUtil.HashToken(rawToken)
	expiresAt := time.Now().Add(s.verifyTTL)

	if err := s.store.SavePasswordResetToken(ctx, acc.ID(), hashedToken, expiresAt); err != nil {
		return err
	}

	return s.emailSvc.SendPasswordResetEmail(email, rawToken)
}

func (s *AuthService) ResetPassword(ctx context.Context, rawToken, newPassword string) error {
	hashedToken := tokenUtil.HashToken(rawToken)

	accID, err := s.store.GetAccountIDByPasswordResetToken(ctx, hashedToken)
	if err != nil {
		return errors.New("invalid or expired reset token")
	}

	_ = s.store.DeletePasswordResetToken(ctx, hashedToken)

	acc, err := s.accountRepo.GetByID(ctx, accID)
	if err != nil {
		return err
	}

	newHash, err := s.hasher.Hash(newPassword)
	if err != nil {
		return err
	}

	acc.UpdatePasswordHash(&newHash)
	if err := s.accountRepo.Update(ctx, acc); err != nil {
		return err
	}

	return s.store.RevokeAllByAccountID(ctx, acc.ID())
}

func (s *AuthService) ChangePassword(ctx context.Context, accID account.AccountID, currentPassword, newPassword string) error {
	acc, err := s.accountRepo.GetByID(ctx, accID)
	if err != nil {
		return err
	}

	match, err := s.hasher.Compare(*acc.PasswordHash(), currentPassword)
	if err != nil || !match {
		return auth.ErrInvalidCredentials
	}

	newHash, err := s.hasher.Hash(newPassword)
	if err != nil {
		return err
	}

	acc.UpdatePasswordHash(&newHash)
	if err := s.accountRepo.Update(ctx, acc); err != nil {
		return err
	}

	return s.store.RevokeAllByAccountID(ctx, acc.ID())
}
