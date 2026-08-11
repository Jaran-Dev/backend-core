package http

import (
	"net/http"
	"time"

	"github.com/Jaran-Dev/backend-core/internal/application/auth"
	"github.com/Jaran-Dev/backend-core/internal/domain/account"
	"github.com/gin-gonic/gin"
)

type AuthHandler struct {
	svc *auth.AuthService
}

func NewAuthHandler(svc *auth.AuthService) *AuthHandler {
	return &AuthHandler{svc: svc}
}

func (h *AuthHandler) Register(c *gin.Context) {
	var req struct {
		Email       string `json:"email" binding:"required,email"`
		Password    string `json:"password" binding:"required,min=8"`
		DisplayName string `json:"display_name"`
		UILocale    string `json:"ui_locale"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	u, err := h.svc.Register(c.Request.Context(), req.Email, req.Password, req.DisplayName, req.UILocale)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, gin.H{"user_id": u.AccountID()})
}

func (h *AuthHandler) LoginUser(c *gin.Context) {
	var req struct {
		Email    string `json:"email" binding:"required,email"`
		Password string `json:"password" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	access, refresh, exp, err := h.svc.Login(c.Request.Context(), req.Email, req.Password)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid credentials"})
		return
	}

	h.setRefreshCookie(c, refresh, exp)
	c.JSON(http.StatusOK, gin.H{
		"access_token": access,
		"expires_in":   int(time.Until(exp).Seconds()), // Note: access token exp differs from refresh exp
	})
}

func (h *AuthHandler) Refresh(c *gin.Context) {
	rawRefresh, err := c.Cookie("__Host-refresh_token")
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "missing refresh token"})
		return
	}

	access, newRefresh, exp, err := h.svc.Refresh(c.Request.Context(), rawRefresh)
	if err != nil {
		// Reuse detected: force clear cookie
		c.SetCookie("__Host-refresh_token", "", -1, "/", "", true, true)
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid refresh session"})
		return
	}

	h.setRefreshCookie(c, newRefresh, exp)
	c.JSON(http.StatusOK, gin.H{"access_token": access})
}

func (h *AuthHandler) Logout(c *gin.Context) {
	rawRefresh, err := c.Cookie("__Host-refresh_token")
	if err == nil && rawRefresh != "" {
		_ = h.svc.Logout(c.Request.Context(), rawRefresh)
	}
	c.SetCookie("__Host-refresh_token", "", -1, "/", "", true, true)
	c.JSON(http.StatusOK, gin.H{"message": "logged out"})
}

func (h *AuthHandler) setRefreshCookie(c *gin.Context, token string, expires time.Time) {
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

func (h *AuthHandler) VerifyEmail(c *gin.Context) {
	var req struct {
		Token string `json:"token" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.svc.VerifyEmail(c.Request.Context(), req.Token); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid or expired token"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "email verified successfully"})
}

func (h *AuthHandler) ResendVerification(c *gin.Context) {
	var req struct {
		Email string `json:"email" binding:"required,email"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	_ = h.svc.ResendVerification(c.Request.Context(), req.Email)
	c.JSON(http.StatusOK, gin.H{"message": "if the email exists, a verification link has been sent"})
}

func (h *AuthHandler) ForgotPassword(c *gin.Context) {
	var req struct {
		Email string `json:"email" binding:"required,email"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	_ = h.svc.ForgotPassword(c.Request.Context(), req.Email)
	c.JSON(http.StatusOK, gin.H{"message": "if the email exists, a password reset link has been sent"})
}

func (h *AuthHandler) ResetPassword(c *gin.Context) {
	var req struct {
		Token       string `json:"token" binding:"required"`
		NewPassword string `json:"new_password" binding:"required,min=8"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.svc.ResetPassword(c.Request.Context(), req.Token, req.NewPassword); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "password reset successfully"})
}

func (h *AuthHandler) ChangePassword(c *gin.Context) {
	var req struct {
		CurrentPassword string `json:"current_password" binding:"required"`
		NewPassword     string `json:"new_password" binding:"required,min=8"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	accIDVal, exists := c.Get("account_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}
	accID := accIDVal.(account.AccountID)

	if err := h.svc.ChangePassword(c.Request.Context(), accID, req.CurrentPassword, req.NewPassword); err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid current password"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "password changed successfully. please log in again."})
}
