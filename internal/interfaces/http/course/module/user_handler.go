package modulehttp

import (
	"context"
	"net/http"

	domainCourse "github.com/Jaran-Dev/backend-core/internal/domain/course"
	domainModule "github.com/Jaran-Dev/backend-core/internal/domain/course/module"
	rootHttp "github.com/Jaran-Dev/backend-core/internal/interfaces/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type UserModuleUseCase interface {
	ListModulesByCourse(ctx context.Context, courseID domainCourse.CourseID) ([]*domainModule.Module, error)
}

type UserModuleHandler struct {
	uc UserModuleUseCase
}

func NewStudentModuleHandler(uc UserModuleUseCase) *UserModuleHandler {
	return &UserModuleHandler{uc: uc}
}

func (h *UserModuleHandler) RegisterRoutes(rg *gin.RouterGroup) {
	courses := rg.Group("/courses")
	{
		courses.GET("/:courseId/modules", h.ListByCourse)
	}
}

func (h *UserModuleHandler) ListByCourse(c *gin.Context) {
	rawCourseID, err := uuid.Parse(c.Param("courseId"))
	if err != nil {
		c.Error(rootHttp.NewHTTPError(http.StatusBadRequest, "VALIDATION_ERROR", "Invalid course ID format: must be a valid UUID."))
		return
	}
	courseID := domainCourse.CourseID(rawCourseID)

	var query ListModulesQuery
	if err := c.ShouldBindQuery(&query); err != nil {
		c.Error(err)
		return
	}

	modules, err := h.uc.ListModulesByCourse(c.Request.Context(), courseID)
	if err != nil {
		c.Error(err)
		return
	}

	data := mapToPublicModuleListDTOs(modules)

	response := PublicModuleListResponse{
		Modules: data,
	}
	c.JSON(http.StatusOK, response)
}

func mapToPublicModuleListDTOs(modules []*domainModule.Module) []PublicModuleResponse {
	data := make([]PublicModuleResponse, len(modules))

	for i, m := range modules {
		data[i] = PublicModuleResponse{
			ID:       uuid.UUID(m.ID()),
			Title:    m.Title(),
			Position: int(m.Position()),
		}
	}

	return data
}
