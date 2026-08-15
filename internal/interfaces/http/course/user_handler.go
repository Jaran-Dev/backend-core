package coursehttp

import (
	"context"
	"net/http"

	domainCourse "github.com/Jaran-Dev/backend-core/internal/domain/course"

	"github.com/gin-gonic/gin"
)

type StudentCourseUseCase interface {
	GetPublishedCourse(ctx context.Context, slug domainCourse.Slug, preferredLocale domainCourse.Locale) (*CourseReadModel, error)
	ListPublishedCourses(ctx context.Context, params domainCourse.QueryParams) ([]*CourseReadModel, int64, error)
}

type StudentCourseHandler struct {
	uc StudentCourseUseCase
}

func NewStudentCourseHandler(uc StudentCourseUseCase) *StudentCourseHandler {
	return &StudentCourseHandler{uc: uc}
}

func (h *StudentCourseHandler) RegisterRoutes(rg *gin.RouterGroup) {
	courses := rg.Group("/courses")
	{
		courses.GET("", h.List)
		courses.GET("/:slug", h.GetBySlug)
	}
}

// GetBySlug handles GET /api/v1/courses/:slug?locale=ckb
func (h *StudentCourseHandler) GetBySlug(c *gin.Context) {
	slug, err := domainCourse.NewSlug(c.Param("slug"))
	if err != nil {
		c.Error(err)
		return
	}

	var query GetCourseBySlugQuery
	if err := c.ShouldBindQuery(&query); err != nil {
		c.Error(err)
		return
	}

	preferredLocale := domainCourse.LocaleEnglish
	if query.Locale != "" {
		preferredLocale = domainCourse.Locale(query.Locale)
	}

	readModel, err := h.uc.GetPublishedCourse(c.Request.Context(), slug, preferredLocale)
	if err != nil {
		c.Error(err)
		return
	}

	response := mapReadModelToDTO(readModel)
	c.JSON(http.StatusOK, response)
}

// List handles GET /api/v1/courses
func (h *StudentCourseHandler) List(c *gin.Context) {
	c.JSON(http.StatusNotImplemented, gin.H{"message": "List endpoint not implemented yet"})
}

func mapReadModelToDTO(rm *CourseReadModel) PublicCourseDetailDTO {
	return PublicCourseDetailDTO{
		ID:             rm.ID,
		Slug:           rm.Slug,
		CoverURL:       rm.CoverURL,
		Title:          rm.Title,
		Description:    rm.Description,
		ResolvedLocale: rm.ResolvedLocale,
	}
}
