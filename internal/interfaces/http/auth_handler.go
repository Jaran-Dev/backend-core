package http

import (
	"net/http"
	"time"

	"github.com/Jaran-Dev/backend-core/internal/application/auth"
	domainAuth "github.com/Jaran-Dev/backend-core/internal/domain/auth"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type AuthHandler struct {
	svc *auth.AuthService
}

func NewAuthHandler(svc *auth.AuthService) *AuthHandler {
	return &AuthHandler{svc: svc}
}

func extractMeta(c *gin.Context) domainAuth.SessionMetadata {
	return domainAuth.SessionMetadata{
		IPAddress:  c.ClientIP(),
		UserAgent:  c.GetHeader("User-Agent"),
		DeviceType: "Unknown", // In the future, parse UserAgent to determine "Mobile", "Desktop", etc.
	}
}

func (h *AuthHandler) Register(c *gin.Context) {
	var req RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.Error(err)
		return
	}

	u, err := h.svc.Register(c.Request.Context(), req.Email, req.Password, req.DisplayName)
	if err != nil {
		c.Error(err)
		return
	}

	c.JSON(http.StatusCreated, gin.H{"account_id": u.AccountID()})
}

func (h *AuthHandler) Login(c *gin.Context) {
	var req LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.Error(err)
		return
	}

	meta := extractMeta(c)
	access, refresh, exp, err := h.svc.Login(c.Request.Context(), req.Email, req.Password, meta)
	if err != nil {
		c.Error(err)
		return
	}

	SetRefreshCookie(c, refresh, exp)
	c.JSON(http.StatusOK, gin.H{
		"access_token": access,
		"expires_in":   int(time.Until(exp).Seconds()),
	})
}

func (h *AuthHandler) Refresh(c *gin.Context) {
	rawRefresh, err := c.Cookie("__Host-refresh_token")
	if err != nil {
		c.Error(NewHTTPError(http.StatusUnauthorized, "UNAUTHORIZED", "missing refresh token"))
		return
	}

	meta := extractMeta(c)
	access, newRefresh, exp, err := h.svc.Refresh(c.Request.Context(), rawRefresh, meta)
	if err != nil {
		c.SetCookie("__Host-refresh_token", "", -1, "/", "", true, true)
		c.Error(err)
		return
	}

	SetRefreshCookie(c, newRefresh, exp)
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

func (h *AuthHandler) ListSessions(c *gin.Context) {
	accID, err := GetAccountIDFromContext(c)
	if err != nil {
		c.Error(err)
		return
	}

	sessions, err := h.svc.ListSessions(c.Request.Context(), accID)
	if err != nil {
		c.Error(NewHTTPError(http.StatusInternalServerError, "INTERNAL_ERROR", "failed to fetch sessions"))
		return
	}

	c.JSON(http.StatusOK, gin.H{"sessions": sessions})
}

func (h *AuthHandler) RevokeSession(c *gin.Context) {
	accID, err := GetAccountIDFromContext(c)
	if err != nil {
		c.Error(err)
		return
	}

	sessionIDStr := c.Param("id")
	sessionID, err := uuid.Parse(sessionIDStr)
	if err != nil {
		c.Error(NewHTTPError(http.StatusBadRequest, "VALIDATION_ERROR", "invalid session id format"))
		return
	}

	if err := h.svc.RevokeSession(c.Request.Context(), accID, sessionID); err != nil {
		c.Error(NewHTTPError(http.StatusInternalServerError, "INTERNAL_ERROR", "failed to revoke session"))
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "session revoked successfully"})
}

func (h *AuthHandler) VerifyEmail(c *gin.Context) {
	var req VerifyEmailRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.Error(err)
		return
	}

	if err := h.svc.VerifyEmail(c.Request.Context(), req.Token); err != nil {
		c.Error(NewHTTPError(http.StatusBadRequest, "INVALID_TOKEN", "invalid or expired token"))
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "email verified successfully"})
}

func (h *AuthHandler) ResendVerification(c *gin.Context) {
	var req ResendVerificationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.Error(err)
		return
	}

	_ = h.svc.ResendVerification(c.Request.Context(), req.Email)
	c.JSON(http.StatusOK, gin.H{"message": "if the email exists, a verification link has been sent"})
}

func (h *AuthHandler) ForgotPassword(c *gin.Context) {
	var req ForgotPasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.Error(err)
		return
	}

	_ = h.svc.ForgotPassword(c.Request.Context(), req.Email)
	c.JSON(http.StatusOK, gin.H{"message": "if the email exists, a password reset link has been sent"})
}

func (h *AuthHandler) ResetPassword(c *gin.Context) {
	var req ResetPasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.Error(err)
		return
	}

	if err := h.svc.ResetPassword(c.Request.Context(), req.Token, req.NewPassword); err != nil {
		c.Error(NewHTTPError(http.StatusBadRequest, "INVALID_TOKEN", "invalid or expired token"))
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "password reset successfully"})
}

func (h *AuthHandler) ChangePassword(c *gin.Context) {
	var req ChangePasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.Error(err)
		return
	}

	accID, err := GetAccountIDFromContext(c)
	if err != nil {
		c.Error(err)
		return
	}

	if err := h.svc.ChangePassword(c.Request.Context(), accID, req.CurrentPassword, req.NewPassword); err != nil {
		c.Error(NewHTTPError(http.StatusUnauthorized, "INVALID_CREDENTIALS", "invalid current password"))
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "password changed successfully. please log in again."})
}
