package http

import (
	"errors"
	"net/http"

	domainAccount "github.com/Jaran-Dev/backend-core/internal/domain/account"
	domainAuth "github.com/Jaran-Dev/backend-core/internal/domain/auth"

	"github.com/gin-gonic/gin"
)

type AuthErrorMapper struct{}

func (m *AuthErrorMapper) MapError(c *gin.Context, err error) bool {
	switch {
	case errors.Is(err, domainAccount.ErrAccountAlreadyExists):
		c.JSON(http.StatusConflict, APIErrorResponse{Code: "ACCOUNT_EXISTS", Message: "Account with this email already exists."})
		return true

	case errors.Is(err, domainAccount.ErrAccountNotFound):
		c.JSON(http.StatusNotFound, APIErrorResponse{Code: "ACCOUNT_NOT_FOUND", Message: "Account not found."})
		return true

	case errors.Is(err, domainAuth.ErrInvalidCredentials):
		c.JSON(http.StatusUnauthorized, APIErrorResponse{Code: "INVALID_CREDENTIALS", Message: "Invalid email or password."})
		return true

	case errors.Is(err, domainAccount.ErrAccountLocked):
		c.JSON(http.StatusLocked, APIErrorResponse{Code: "ACCOUNT_LOCKED", Message: "Account is locked due to too many failed attempts."})
		return true

	case errors.Is(err, domainAccount.ErrAccountBanned):
		c.JSON(http.StatusForbidden, APIErrorResponse{Code: "ACCOUNT_BANNED", Message: "Account is banned."})
		return true

	case errors.Is(err, domainAuth.ErrTokenNotFound), errors.Is(err, domainAuth.ErrTokenExpired), errors.Is(err, domainAuth.ErrRefreshTokenReused):
		c.JSON(http.StatusUnauthorized, APIErrorResponse{Code: "INVALID_REFRESH_SESSION", Message: "Invalid or expired refresh session."})
		return true
	}

	return false
}
