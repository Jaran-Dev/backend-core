package jwt

import (
	"crypto/rsa"
	"fmt"
	"time"

	"github.com/Jaran-Dev/backend-core/internal/domain/account"
	"github.com/Jaran-Dev/backend-core/internal/domain/auth"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

type Config struct {
	AccessTokenTTL       time.Duration
	VerificationTokenTTL time.Duration
	PrivateKey           *rsa.PrivateKey
	PublicKey            *rsa.PublicKey
}

type TokenService struct {
	cfg Config
}

func NewTokenService(cfg Config) *TokenService {
	return &TokenService{cfg: cfg}
}

func (s *TokenService) GenerateAccessToken(accID account.AccountID, role auth.Role) (string, time.Time, error) {
	now := time.Now()
	exp := now.Add(s.cfg.AccessTokenTTL)

	claims := jwt.MapClaims{
		"sub":  accID.String(),
		"role": string(role),
		"iat":  now.Unix(),
		"exp":  exp.Unix(),
		"ttyp": "access",
	}
	token := jwt.NewWithClaims(jwt.SigningMethodRS256, claims)
	tokenString, err := token.SignedString(s.cfg.PrivateKey)
	if err != nil {
		return "", time.Time{}, fmt.Errorf("failed to sign access token: %w", err)
	}
	return tokenString, exp, nil
}

func (s *TokenService) ValidateAccessToken(tokenString string) (*auth.AccessTokenClaims, error) {
	claims, err := s.parseToken(tokenString, "access")
	if err != nil {
		return nil, auth.ErrInvalidCredentials
	}

	uid, _ := uuid.Parse(claims["sub"].(string))
	roleStr, _ := claims["role"].(string)

	return &auth.AccessTokenClaims{
		AccountID: account.AccountID(uid),
		Role:      auth.Role(roleStr),
	}, nil
}

func (s *TokenService) GenerateEmailVerificationToken(accID account.AccountID) (string, error) {
	now := time.Now()
	exp := now.Add(s.cfg.VerificationTokenTTL)

	claims := jwt.MapClaims{
		"sub":  accID.String(),
		"iat":  now.Unix(),
		"exp":  exp.Unix(),
		"ttyp": "email_verify",
	}
	token := jwt.NewWithClaims(jwt.SigningMethodRS256, claims)
	return token.SignedString(s.cfg.PrivateKey)
}

func (s *TokenService) ValidateEmailVerificationToken(tokenString string) (account.AccountID, error) {
	claims, err := s.parseToken(tokenString, "email_verify")
	if err != nil {
		return account.AccountID(uuid.Nil), err
	}

	uid, _ := uuid.Parse(claims["sub"].(string))
	return account.AccountID(uid), nil
}

func (s *TokenService) parseToken(tokenString, expectedType string) (jwt.MapClaims, error) {
	parsed, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodRSA); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return s.cfg.PublicKey, nil
	})
	if err != nil || !parsed.Valid {
		return nil, auth.ErrInvalidCredentials
	}

	claims, ok := parsed.Claims.(jwt.MapClaims)
	if !ok {
		return nil, auth.ErrInvalidCredentials
	}

	ttyp, ok := claims["ttyp"].(string)
	if !ok || ttyp != expectedType {
		return nil, auth.ErrInvalidCredentials
	}

	return claims, nil
}
