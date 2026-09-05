package http

import (
	"net/http"

	"github.com/Jaran-Dev/backend-core/internal/domain/account"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

func GetAccountIDFromContext(c *gin.Context) (account.AccountID, error) {
	val, exists := c.Get("account_id")
	if !exists {
		return account.AccountID(uuid.Nil), NewHTTPError(http.StatusUnauthorized, "UNAUTHORIZED", "missing account_id in context")
	}

	accID, ok := val.(account.AccountID)
	if !ok {
		return account.AccountID(uuid.Nil), NewHTTPError(http.StatusUnauthorized, "UNAUTHORIZED", "invalid account_id type in context")
	}

	return accID, nil
}
