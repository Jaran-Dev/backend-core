package http

import (
	"net/http"
	"strings"

	"github.com/Jaran-Dev/backend-core/internal/domain/auth"
	"github.com/gin-gonic/gin"
)

func AuthMiddleware(tokenSvc auth.TokenService) gin.HandlerFunc {
	return func(c *gin.Context) {
		header := c.GetHeader("Authorization")
		if header == "" || !strings.HasPrefix(header, "Bearer ") {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "missing or malformed auth header"})
			return
		}

		tokenStr := strings.TrimPrefix(header, "Bearer ")
		claims, err := tokenSvc.ValidateAccessToken(tokenStr)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "invalid or expired token"})
			return
		}

		c.Set("user_id", claims.UserID)
		c.Set("role", claims.Role)
		c.Next()
	}
}

func RoleMiddleware(requiredRoles ...auth.Role) gin.HandlerFunc {
	return func(c *gin.Context) {
		roleVal, exists := c.Get("role")
		if !exists {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "forbidden: no role attached to context"})
			return
		}

		userRole := roleVal.(auth.Role)

		if userRole == auth.RoleAdmin {
			c.Next()
			return
		}

		isAllowed := false
		for _, r := range requiredRoles {
			if userRole == r {
				isAllowed = true
				break
			}
		}

		if !isAllowed {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "insufficient permissions"})
			return
		}

		c.Next()
	}
}
