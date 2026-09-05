package http

import (
	"net/http"
	"net/url"
	"strconv"
	"time"

	appOauth "github.com/Jaran-Dev/backend-core/internal/application/oauth"
	domainAuth "github.com/Jaran-Dev/backend-core/internal/domain/auth"
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
		c.Error(NewHTTPError(http.StatusBadRequest, "VALIDATION_ERROR", "invalid provider"))
		return
	}

	redirectURL := c.Query("redirect")
	if redirectURL == "" {
		redirectURL = "/"
	}

	authURL, err := h.svc.InitiateLogin(c.Request.Context(), provider, redirectURL)
	if err != nil {
		c.Error(err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"auth_url": authURL})
}

func (h *OAuthHandler) Callback(c *gin.Context) {
	providerStr := c.Param("provider")
	provider := domainOauth.Provider(providerStr)
	if !provider.IsValid() {
		c.Error(NewHTTPError(http.StatusBadRequest, "VALIDATION_ERROR", "invalid provider"))
		return
	}

	var req OAuthCallbackQuery
	if err := c.ShouldBindQuery(&req); err != nil {
		c.Error(err)
		return
	}

	meta := domainAuth.SessionMetadata{
		IPAddress:  c.ClientIP(),
		UserAgent:  c.GetHeader("User-Agent"),
		DeviceType: "Unknown",
	}

	result, err := h.svc.HandleCallback(c.Request.Context(), provider, req.Code, req.State, meta)
	if err != nil {
		c.Error(err)
		return
	}

	h.setRefreshCookie(c, result.RefreshToken, result.ExpiresAt)

	fragment := url.Values{}
	fragment.Set("access_token", result.AccessToken)
	fragment.Set("token_type", "Bearer")
	fragment.Set("expires_in", strconv.Itoa(int(time.Until(result.ExpiresAt).Seconds())))

	redirectTo, parseErr := url.Parse(result.RedirectURL)
	if parseErr != nil || redirectTo.String() == "" {
		c.Error(NewHTTPError(http.StatusInternalServerError, "INTERNAL_ERROR", "invalid redirect url"))
		return
	}

	if redirectTo.Fragment != "" {
		redirectTo.Fragment += "&"
	}
	redirectTo.Fragment += fragment.Encode()

	c.Redirect(http.StatusFound, redirectTo.String())
}

func (h *OAuthHandler) Link(c *gin.Context) {
	providerStr := c.Param("provider")
	provider := domainOauth.Provider(providerStr)
	if !provider.IsValid() {
		c.Error(NewHTTPError(http.StatusBadRequest, "VALIDATION_ERROR", "invalid provider"))
		return
	}

	var req OAuthLinkQuery
	if err := c.ShouldBindQuery(&req); err != nil {
		c.Error(err)
		return
	}

	accID, err := GetAccountIDFromContext(c)
	if err != nil {
		c.Error(err)
		return
	}

	if err := h.svc.LinkAccount(c.Request.Context(), accID, provider, req.Code, req.State); err != nil {
		c.Error(err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "account linked successfully"})
}

func (h *OAuthHandler) Unlink(c *gin.Context) {
	providerStr := c.Param("provider")
	provider := domainOauth.Provider(providerStr)
	if !provider.IsValid() {
		c.Error(NewHTTPError(http.StatusBadRequest, "VALIDATION_ERROR", "invalid provider"))
		return
	}

	accID, err := GetAccountIDFromContext(c)
	if err != nil {
		c.Error(err)
		return
	}

	if err := h.svc.UnlinkAccount(c.Request.Context(), accID, provider); err != nil {
		c.Error(err)
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
		"",
		true,
		true,
	)
}
