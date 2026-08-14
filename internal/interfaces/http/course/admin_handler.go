package coursehttp

import (
	"context"
	"errors"
	"net/http"
	"time"

	"github.com/Jaran-Dev/backend-core/internal/domain/account"
	domainCourse "github.com/Jaran-Dev/backend-core/internal/domain/course"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type AdminCourseUseCase interface {
	Create(ctx context.Context, slugStr, coverURL string, adminID account.AccountID) (*domainCourse.Course, error)
}

type AdminCourseHandler struct {
	uc AdminCourseUseCase
}

func NewAdminCourseHandler(uc AdminCourseUseCase) *AdminCourseHandler {
	return &AdminCourseHandler{uc: uc}
}

func (h *AdminCourseHandler) RegisterRoutes(rg *gin.RouterGroup) {
	courses := rg.Group("/courses")
	{
		courses.POST("", h.Create)
		// Future routes:
		// courses.GET("", h.List)
		// courses.PATCH("/:id", h.Update)
		// courses.POST("/:id/translations", h.UpsertTranslation)
	}
}

// Create handles POST /api/v1/admin/courses
func (h *AdminCourseHandler) Create(c *gin.Context) {
	var req CreateCourseRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.Error(err)
		return
	}

	adminID, err := getAdminIDFromContext(c)
	if err != nil {
		c.Error(err)
		return
	}

	course, err := h.uc.Create(c.Request.Context(), req.Slug, req.CoverURL, adminID)
	if err != nil {
		c.Error(err)
		return
	}

	response := mapToCourseResponse(course)

	c.JSON(http.StatusCreated, response)
}

func getAdminIDFromContext(c *gin.Context) (account.AccountID, error) {
	idVal, exists := c.Get("user_id")
	if !exists {
		return account.AccountID(uuid.Nil), errors.New("UNAUTHORIZED: missing user_id in context")
	}

	switch v := idVal.(type) {
	case uuid.UUID:
		return account.AccountID(v), nil
	case string:
		id, err := uuid.Parse(v)
		return account.AccountID(id), err
	default:
		return account.AccountID(uuid.Nil), errors.New("UNAUTHORIZED: invalid user_id type in context")
	}
}

func mapToCourseResponse(c *domainCourse.Course) CourseResponse {
	var transResp []TranslationResponse
	if c.Translations() != nil {
		transResp = make([]TranslationResponse, len(c.Translations()))
		for i, t := range c.Translations() {
			transResp[i] = TranslationResponse{
				Locale:      string(t.Locale()),
				Title:       t.Title(),
				Description: t.Description(),
			}
		}
	}

	return CourseResponse{
		ID:           uuid.UUID(c.ID()),
		Slug:         c.Slug().String(),
		Status:       string(c.Status()),
		CoverURL:     c.CoverURL(),
		CreatedBy:    uuid.UUID(c.CreatedBy()),
		CreatedAt:    c.CreatedAt().Format(time.RFC3339), // Standardized ISO 8601 format for JS
		UpdatedAt:    c.UpdatedAt().Format(time.RFC3339),
		Translations: transResp,
	}
}
