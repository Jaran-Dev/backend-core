package http

import (
	"net/http"
	"time"

	appOauth "github.com/Jaran-Dev/backend-core/internal/application/oauth"
	domainAccount "github.com/Jaran-Dev/backend-core/internal/domain/account"
	domainOauth "github.com/Jaran-Dev/backend-core/internal/domain/oauth"
	"github.com/gin-gonic/gin"
)

type OAuthHandler struct {
	svc *appOauth.OAuthService
}

func NewOAuthHandler(svc *appOauth.OAuthService) *OAuthHandler {
	return &OAuthHandler{svc: svc}
}

func (h *OAuthHandler) InitiateLogin(c *gin.Context) {
	providerStr := c.Param("provider")
	provider := domainOauth.Provider(providerStr)
	if !provider.IsValid() {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid provider"})
		return
	}

	redirectURL := c.Query("redirect")
	if redirectURL == "" {
		redirectURL = "/" // Default redirect
	}

	authURL, err := h.svc.InitiateLogin(c.Request.Context(), provider, redirectURL)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"auth_url": authURL})
}

func (h *OAuthHandler) Callback(c *gin.Context) {
	providerStr := c.Param("provider")
	provider := domainOauth.Provider(providerStr)
	if !provider.IsValid() {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid provider"})
		return
	}

	code := c.Query("code")
	state := c.Query("state")
	if code == "" || state == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "missing code or state"})
		return
	}

	access, refresh, exp, err := h.svc.HandleCallback(c.Request.Context(), provider, code, state)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}

	h.setRefreshCookie(c, refresh, exp)

	// In a real SPA, you'd redirect to the frontend with the token in the URL fragment
	// e.g., /auth/callback#access_token=...
	c.JSON(http.StatusOK, gin.H{
		"access_token": access,
		"expires_in":   int(time.Until(exp).Seconds()),
	})
}

func (h *OAuthHandler) Link(c *gin.Context) {
	providerStr := c.Param("provider")
	provider := domainOauth.Provider(providerStr)
	if !provider.IsValid() {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid provider"})
		return
	}

	code := c.Query("code")
	state := c.Query("state")
	if code == "" || state == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "missing code or state"})
		return
	}

	accIDVal, exists := c.Get("account_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}
	accID := accIDVal.(domainAccount.AccountID) // Note: Adjust type import based on your context setup

	if err := h.svc.LinkAccount(c.Request.Context(), accID, provider, code, state); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "account linked successfully"})
}

func (h *OAuthHandler) Unlink(c *gin.Context) {
	providerStr := c.Param("provider")
	provider := domainOauth.Provider(providerStr)
	if !provider.IsValid() {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid provider"})
		return
	}

	accIDVal, exists := c.Get("account_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}
	accID := accIDVal.(domainAccount.AccountID)

	if err := h.svc.UnlinkAccount(c.Request.Context(), accID, provider); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "account unlinked successfully"})
}

func (h *OAuthHandler) setRefreshCookie(c *gin.Context, token string, expires time.Time) {
	c.SetSameSite(http.SameSiteLaxMode)
	c.SetCookie(
		"__Host-refresh_token",
		token,
		int(time.Until(expires).Seconds()),
		"/",
		"",   // Empty domain -> host-only
		true, // Secure (requires HTTPS in prod)
		true, // HttpOnly
	)
}
