package http

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

func SetRefreshCookie(c *gin.Context, token string, expires time.Time) {
	c.SetSameSite(http.SameSiteLaxMode)
	c.SetCookie(
		"__Host-refresh_token",
		token,
		int(time.Until(expires).Seconds()),
		"/",
		"",
		true,
		true,
	)
}
