package coursehttp

import (
	"context"
	"net/http"
	"time"

	"github.com/Jaran-Dev/backend-core/internal/domain/account"
	domainCourse "github.com/Jaran-Dev/backend-core/internal/domain/course"
	rootHttp "github.com/Jaran-Dev/backend-core/internal/interfaces/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type AdminCourseUseCase interface {
	Create(ctx context.Context, slugStr, coverURL string, adminID account.AccountID) (*domainCourse.Course, error)
	UpdateMetadata(ctx context.Context, courseID domainCourse.CourseID, slug *string, coverURL *string) (*domainCourse.Course, error)
	UpsertTranslation(ctx context.Context, courseID domainCourse.CourseID, locale domainCourse.Locale, title, description string) (*domainCourse.Course, error)
	Publish(ctx context.Context, courseID domainCourse.CourseID) (*domainCourse.Course, error)
	Archive(ctx context.Context, courseID domainCourse.CourseID) (*domainCourse.Course, error)
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
		courses.PATCH("/:id/status", h.ChangeStatus)
	}
}

// Create handles POST /api/v1/admin/courses
func (h *AdminCourseHandler) Create(c *gin.Context) {
	var req CreateCourseRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.Error(err)
		return
	}

	accountID, err := rootHttp.GetAccountIDFromContext(c)
	if err != nil {
		c.Error(err)
		return
	}

	course, err := h.uc.Create(c.Request.Context(), req.Slug, req.CoverURL, accountID)
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

// UpsertTranslation handles POST /api/v1/admin/courses/:id/translations
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

// ChangeStatus handles PATCH /api/v1/admin/courses/:id/status
func (h *AdminCourseHandler) ChangeStatus(c *gin.Context) {
	rawID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.Error(rootHttp.NewHTTPError(http.StatusBadRequest, "VALIDATION_ERROR", "Invalid course ID format: must be a valid UUID."))
		return
	}
	courseID := domainCourse.CourseID(rawID)

	var req UpdateStatusRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.Error(err)
		return
	}

	var updatedCourse *domainCourse.Course
	switch req.Action {
	case "publish":
		updatedCourse, err = h.uc.Publish(c.Request.Context(), courseID)
	case "archive":
		updatedCourse, err = h.uc.Archive(c.Request.Context(), courseID)
	}

	if err != nil {
		c.Error(err)
		return
	}

	response := mapToCourseResponse(updatedCourse)
	c.JSON(http.StatusOK, response)
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
