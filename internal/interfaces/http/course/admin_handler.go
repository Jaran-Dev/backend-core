package coursehttp

import (
	"context"
	"errors"
	"net/http"
	"time"

	domainAdmin "github.com/Jaran-Dev/backend-core/internal/domain/admin"
	domainCourse "github.com/Jaran-Dev/backend-core/internal/domain/course"
	rootHttp "github.com/Jaran-Dev/backend-core/internal/interfaces/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type AdminCourseUseCase interface {
	Create(ctx context.Context, slugStr, coverURL string, adminID domainAdmin.AdminID) (*domainCourse.Course, error)
	UpdateMetadata(ctx context.Context, courseID domainCourse.CourseID, slug *string, coverURL *string) (*domainCourse.Course, error)
	UpsertTranslation(ctx context.Context, courseID domainCourse.CourseID, locale domainCourse.Locale, title, description string) (*domainCourse.Course, error)
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
		courses.PATCH("/:id", h.Update)
		courses.POST("/:id/translations", h.UpsertTranslation)
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

// Update handles PATCH /api/v1/admin/courses/:id
func (h *AdminCourseHandler) Update(c *gin.Context) {
	rawID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.Error(rootHttp.NewHTTPError(http.StatusBadRequest, "VALIDATION_ERROR", "Invalid course ID format: must be a valid UUID."))
		return
	}

	courseID := domainCourse.CourseID(rawID)

	var req UpdateCourseRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.Error(err)
		return
	}

	updatedCourse, err := h.uc.UpdateMetadata(c.Request.Context(), courseID, req.Slug, req.CoverURL)
	if err != nil {
		c.Error(err)
		return
	}

	response := mapToCourseResponse(updatedCourse)
	c.JSON(http.StatusOK, response)
}

func (h *AdminCourseHandler) UpsertTranslation(c *gin.Context) {
	rawID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.Error(rootHttp.NewHTTPError(http.StatusBadRequest, "VALIDATION_ERROR", "Invalid course ID format: must be a valid UUID."))
		return
	}
	courseID := domainCourse.CourseID(rawID)

	var req UpsertTranslationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.Error(err)
		return
	}

	locale := domainCourse.Locale(req.Locale)

	updatedCourse, err := h.uc.UpsertTranslation(c.Request.Context(), courseID, locale, req.Title, req.Description)
	if err != nil {
		c.Error(err)
		return
	}

	response := mapToCourseResponse(updatedCourse)
	c.JSON(http.StatusOK, response)
}

func getAdminIDFromContext(c *gin.Context) (domainAdmin.AdminID, error) {
	idVal, exists := c.Get("user_id")
	if !exists {
		return domainAdmin.AdminID(uuid.Nil), errors.New("UNAUTHORIZED: missing user_id in context")
	}

	var rawUUID uuid.UUID
	switch v := idVal.(type) {
	case uuid.UUID:
		rawUUID = v
	case string:
		parsed, err := uuid.Parse(v)
		if err != nil {
			return domainAdmin.AdminID(uuid.Nil), errors.New("UNAUTHORIZED: invalid user_id format in context")
		}
		rawUUID = parsed
	default:
		return domainAdmin.AdminID(uuid.Nil), errors.New("UNAUTHORIZED: invalid user_id type in context")
	}

	return domainAdmin.AdminID(rawUUID), nil
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
		CreatedBy:    uuid.UUID(c.CreatedBy()), // Translate back to raw UUID for JSON serialization
		CreatedAt:    c.CreatedAt().Format(time.RFC3339),
		UpdatedAt:    c.UpdatedAt().Format(time.RFC3339),
		Translations: transResp,
	}
}
