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
	AccessTokenTTL time.Duration
	PrivateKey     *rsa.PrivateKey
	PublicKey      *rsa.PublicKey
}

type TokenService struct {
	cfg Config
}

func NewTokenService(cfg Config) *TokenService {
	return &TokenService{cfg: cfg}
}

func (s *TokenService) GenerateAccessToken(userID uuid.UUID, role auth.Role) (string, time.Time, error) {
	now := time.Now()
	exp := now.Add(s.cfg.AccessTokenTTL)

	claims := jwt.MapClaims{
		"sub":  userID.String(),
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
	parsed, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodRSA); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return s.cfg.PublicKey, nil
	})
	if err != nil {
		return nil, auth.ErrInvalidCredentials
	}
	claims, ok := parsed.Claims.(jwt.MapClaims)
	if !ok || !parsed.Valid {
		return nil, auth.ErrInvalidCredentials
	}

	uidStr, _ := claims["sub"].(string)
	uid, err := uuid.Parse(uidStr)
	if err != nil {
		return nil, auth.ErrInvalidCredentials
	}

	roleStr, _ := claims["role"].(string)
	return &auth.AccessTokenClaims{
		AccountID: account.AccountID(uid),
		Role:      auth.Role(roleStr),
	}, nil
}
