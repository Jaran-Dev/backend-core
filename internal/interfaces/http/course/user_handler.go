package coursehttp

import (
    "context"

    domainCourse "github.com/Jaran-Dev/backend-core/internal/domain/course"

    "github.com/gin-gonic/gin"
)

type UserCourseUseCase interface {
    GetPublishedCourse(ctx context.Context, slug domainCourse.Slug, preferredLocale domainCourse.Locale) (*domainCourse.Course, domainCourse.Locale, error)
    ListPublishedCourses(ctx context.Context, params domainCourse.QueryParams) ([]*domainCourse.Course, int64, error)
}

type UserCourseHandler struct {
    uc UserCourseUseCase
}

func NewStudentCourseHandler(uc UserCourseUseCase) *UserCourseHandler {
    return &UserCourseHandler{uc: uc}
}

func (h *UserCourseHandler) RegisterRoutes(rg *gin.RouterGroup) {
    courses := rg.Group("/courses")
    {
        courses.GET("", h.List) 
        courses.GET("/:slug", h.GetBySlug) 
    }
}

// List handles GET /api/v1/courses
func (h *UserCourseHandler) List(c *gin.Context) {
    c.JSON(501, gin.H{"message": "List endpoint not implemented yet"})
}

// GetBySlug handles GET /api/v1/courses/:slug
func (h *UserCourseHandler) GetBySlug(c *gin.Context) {
    c.JSON(501, gin.H{"message": "GetBySlug endpoint not implemented yet"})
}