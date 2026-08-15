package coursehttp

import (
	"context"
	"net/http"

	domainCourse "github.com/Jaran-Dev/backend-core/internal/domain/course"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type StudentCourseUseCase interface {
	GetPublishedCourse(ctx context.Context, slug domainCourse.Slug, preferredLocale domainCourse.Locale) (*domainCourse.Course, domainCourse.Locale, error)
	ListPublishedCourses(ctx context.Context, params domainCourse.QueryParams, locale domainCourse.Locale) ([]*domainCourse.Course, int64, error)
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

	course, resolvedLocale, err := h.uc.GetPublishedCourse(c.Request.Context(), slug, preferredLocale)
	if err != nil {
		c.Error(err)
		return
	}

	response := mapToPublicCourseDetailDTO(course, resolvedLocale)
	c.JSON(http.StatusOK, response)
}

// List handles GET /api/v1/courses?page=1&page_size=10&locale=ckb
func (h *StudentCourseHandler) List(c *gin.Context) {
	var query ListCoursesQuery
	if err := c.ShouldBindQuery(&query); err != nil {
		c.Error(err)
		return
	}

	query.SanitizeAndDefault()

	params := domainCourse.QueryParams{
		Page:     query.Page,
		PageSize: query.PageSize,
		Status:   string(domainCourse.CourseStatusPublished),
	}

	preferredLocale := domainCourse.Locale(query.Locale)

	courses, total, err := h.uc.ListPublishedCourses(c.Request.Context(), params, preferredLocale)
	if err != nil {
		c.Error(err)
		return
	}

	data := mapToPublicCourseListDTOs(courses)

	response := NewPaginatedResponse(data, total, query.Page, query.PageSize)
	c.JSON(http.StatusOK, response)
}

func mapToPublicCourseListDTOs(courses []*domainCourse.Course) []PublicCourseListItemDTO {
	data := make([]PublicCourseListItemDTO, len(courses))

	for i, course := range courses {
		title := ""
		if len(course.Translations()) > 0 {
			title = course.Translations()[0].Title()
		}

		data[i] = PublicCourseListItemDTO{
			ID:       uuid.UUID(course.ID()),
			Slug:     course.Slug().String(),
			CoverURL: course.CoverURL(),
			Title:    title,
		}
	}

	return data
}

func mapToPublicCourseDetailDTO(c *domainCourse.Course, resolvedLocale domainCourse.Locale) PublicCourseDetailDTO {
	var title, description string

	if c.Translations() != nil && len(c.Translations()) > 0 {
		title = c.Translations()[0].Title()
		description = c.Translations()[0].Description()
	}

	return PublicCourseDetailDTO{
		ID:             uuid.UUID(c.ID()),
		Slug:           c.Slug().String(),
		CoverURL:       c.CoverURL(),
		Title:          title,
		Description:    description,
		ResolvedLocale: string(resolvedLocale),
	}
}
