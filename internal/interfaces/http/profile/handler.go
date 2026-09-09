package profilehttp

import (
	"net/http"
	"strings"
	"time"

	appProfile "github.com/Jaran-Dev/backend-core/internal/application/profile"
	rootHttp "github.com/Jaran-Dev/backend-core/internal/interfaces/http"

	"github.com/gin-gonic/gin"
)

type ProfileHandler struct {
	svc *appProfile.ProfileService
}

func NewProfileHandler(svc *appProfile.ProfileService) *ProfileHandler {
	return &ProfileHandler{svc: svc}
}

func (h *ProfileHandler) RegisterRoutes(rg *gin.RouterGroup) {
	me := rg.Group("/me")
	{
		me.GET("/profile", h.GetProfile)
		me.PATCH("/profile", h.UpdateProfile)
		me.POST("/profile/avatar", h.UploadAvatar)
		me.POST("/profile/cover", h.UploadCover)
		me.GET("/preferences", h.GetPreferences)
		me.PATCH("/preferences", h.UpdatePreferences)
	}
}

func (h *ProfileHandler) GetProfile(c *gin.Context) {
	accID, err := rootHttp.GetAccountIDFromContext(c)
	if err != nil {
		c.Error(err)
		return
	}

	u, a, err := h.svc.GetProfile(c.Request.Context(), accID)
	if err != nil {
		c.Error(err)
		return
	}

	dob := ""
	if !time.Time(u.DateOfBirth()).IsZero() {
		dob = u.DateOfBirth().Time().Format("2006-01-02")
	}

	response := ProfileResponse{
		AccountID:   u.AccountID().String(),
		Email:       a.Email(),
		DisplayName: u.DisplayName(),
		Username:    string(u.Username()),
		Bio:         string(u.Bio()),
		Headline:    string(u.Headline()),
		Pronouns:    string(u.Pronouns()),
		Location:    string(u.Location()),
		DateOfBirth: dob,
		AvatarURL:   u.Avatar().URL(),
		CoverURL:    u.Cover().URL(),
	}

	c.JSON(http.StatusOK, response)
}

func (h *ProfileHandler) UpdateProfile(c *gin.Context) {
	var req UpdateProfileRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.Error(err)
		return
	}

	accID, err := rootHttp.GetAccountIDFromContext(c)
	if err != nil {
		c.Error(err)
		return
	}

	var dob *time.Time
	if req.DateOfBirth != nil && *req.DateOfBirth != "" {
		t, err := time.Parse("2006-01-02", *req.DateOfBirth)
		if err != nil {
			c.Error(rootHttp.NewHTTPError(http.StatusBadRequest, "VALIDATION_ERROR", "invalid date_of_birth format, must be YYYY-MM-DD"))
			return
		}
		dob = &t
	}

	params := appProfile.UpdateProfileParams{
		DisplayName: req.DisplayName,
		Username:    req.Username,
		Bio:         req.Bio,
		Headline:    req.Headline,
		Pronouns:    req.Pronouns,
		Location:    req.Location,
		DateOfBirth: dob,
	}

	if err := h.svc.UpdateProfile(c.Request.Context(), accID, params); err != nil {
		c.Error(err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "profile updated successfully"})
}

func (h *ProfileHandler) UploadAvatar(c *gin.Context) {
	h.uploadImage(c, "avatar")
}

func (h *ProfileHandler) UploadCover(c *gin.Context) {
	h.uploadImage(c, "cover")
}

func (h *ProfileHandler) uploadImage(c *gin.Context, imageType string) {
	accID, err := rootHttp.GetAccountIDFromContext(c)
	if err != nil {
		c.Error(err)
		return
	}

	file, _, err := c.Request.FormFile("file")
	if err != nil {
		c.Error(rootHttp.NewHTTPError(http.StatusBadRequest, "VALIDATION_ERROR", "no file uploaded"))
		return
	}
	defer file.Close()

	contentType := c.GetHeader("Content-Type")
	if !strings.HasPrefix(contentType, "image/") {
		c.Error(rootHttp.NewHTTPError(http.StatusBadRequest, "VALIDATION_ERROR", "file must be an image"))
		return
	}

	var svcErr error
	if imageType == "avatar" {
		svcErr = h.svc.UploadAvatar(c.Request.Context(), accID, file, contentType)
	} else {
		svcErr = h.svc.UploadCover(c.Request.Context(), accID, file, contentType)
	}

	if svcErr != nil {
		c.Error(svcErr)
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": imageType + " uploaded successfully"})
}

func (h *ProfileHandler) GetPreferences(c *gin.Context) {
	accID, err := rootHttp.GetAccountIDFromContext(c)
	if err != nil {
		c.Error(err)
		return
	}

	p, err := h.svc.GetPreferences(c.Request.Context(), accID)
	if err != nil {
		c.Error(err)
		return
	}

	response := PreferencesResponse{
		UILocale:             string(p.UILocale()),
		Theme:                string(p.Theme()),
		PlaybackQuality:      string(p.PlaybackQuality()),
		AutoplayNext:         p.AutoplayNext(),
		SubtitleLanguage:     string(p.SubtitleLanguage()),
		SubtitlesEnabled:     p.SubtitlesEnabled(),
		SkipIntro:            p.SkipIntro(),
		ReducedMotion:        p.ReducedMotion(),
		EmailDigestFrequency: string(p.EmailDigestFrequency()),
		NotificationSettings: p.NotificationSettings(),
	}

	c.JSON(http.StatusOK, response)
}

func (h *ProfileHandler) UpdatePreferences(c *gin.Context) {
	var req UpdatePreferencesRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.Error(err)
		return
	}

	accID, err := rootHttp.GetAccountIDFromContext(c)
	if err != nil {
		c.Error(err)
		return
	}

	params := appProfile.UpdatePreferencesParams{
		UILocale:             req.UILocale,
		Theme:                req.Theme,
		PlaybackQuality:      req.PlaybackQuality,
		AutoplayNext:         req.AutoplayNext,
		SubtitleLanguage:     req.SubtitleLanguage,
		SubtitlesEnabled:     req.SubtitlesEnabled,
		SkipIntro:            req.SkipIntro,
		ReducedMotion:        req.ReducedMotion,
		EmailDigestFrequency: req.EmailDigestFrequency,
		NotificationSettings: req.NotificationSettings,
	}

	if err := h.svc.UpdatePreferences(c.Request.Context(), accID, params); err != nil {
		c.Error(err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "preferences updated successfully"})
}
