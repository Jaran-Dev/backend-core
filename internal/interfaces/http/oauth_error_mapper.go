package http

import (
	"errors"
	"net/http"

	domainOauth "github.com/Jaran-Dev/backend-core/internal/domain/oauth"

	"github.com/gin-gonic/gin"
)

type OAuthErrorMapper struct{}

func (m *OAuthErrorMapper) MapError(c *gin.Context, err error) bool {
	switch {
	case errors.Is(err, domainOauth.ErrOAuthAccountNotFound):
		c.JSON(http.StatusNotFound, APIErrorResponse{Code: "OAUTH_ACCOUNT_NOT_FOUND", Message: "OAuth account not found."})
		return true

	case errors.Is(err, domainOauth.ErrAccountAlreadyLinked):
		c.JSON(http.StatusConflict, APIErrorResponse{Code: "ACCOUNT_ALREADY_LINKED", Message: "Account is already linked to this provider."})
		return true

	case errors.Is(err, domainOauth.ErrCannotUnlinkLastMethod):
		c.JSON(http.StatusBadRequest, APIErrorResponse{Code: "CANNOT_UNLINK_LAST_METHOD", Message: "Cannot unlink the last authentication method."})
		return true

	case errors.Is(err, domainOauth.ErrInvalidState), errors.Is(err, domainOauth.ErrOAuthCodeExchangeFailed):
		c.JSON(http.StatusBadRequest, APIErrorResponse{Code: "OAUTH_ERROR", Message: "OAuth state is invalid or code exchange failed."})
		return true

	case errors.Is(err, domainOauth.ErrOAuthProviderUnsupported):
		c.JSON(http.StatusBadRequest, APIErrorResponse{Code: "UNSUPPORTED_PROVIDER", Message: "Unsupported OAuth provider."})
		return true
	}

	// Not an oauth error
	return false
}
