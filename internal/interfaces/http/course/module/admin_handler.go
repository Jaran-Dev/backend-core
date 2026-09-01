package modulehttp

import (
	"context"
	"net/http"
	"time"

	domainCourse "github.com/Jaran-Dev/backend-core/internal/domain/course"
	domainModule "github.com/Jaran-Dev/backend-core/internal/domain/course/module"
	rootHttp "github.com/Jaran-Dev/backend-core/internal/interfaces/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type AdminModuleUseCase interface {
	Create(ctx context.Context, courseID domainCourse.CourseID, title string) (*domainModule.Module, error)
	UpdateMetadata(ctx context.Context, moduleID domainModule.ModuleID, title *string) (*domainModule.Module, error)
	Reorder(ctx context.Context, courseID domainCourse.CourseID, moduleIDs []uuid.UUID) ([]*domainModule.Module, error)
	Delete(ctx context.Context, moduleID domainModule.ModuleID) error
}

type AdminModuleHandler struct {
	uc AdminModuleUseCase
}

func NewAdminModuleHandler(uc AdminModuleUseCase) *AdminModuleHandler {
	return &AdminModuleHandler{uc: uc}
}

func (h *AdminModuleHandler) RegisterRoutes(rg *gin.RouterGroup) {
	modules := rg.Group("/modules")
	{
		modules.POST("", h.Create)
		modules.PATCH("/:id", h.Update)
		modules.DELETE("/:id", h.Delete)
	}

	rg.PUT("/courses/:courseId/modules/order", h.Reorder)
}

// Create handles POST /api/v1/admin/modules
func (h *AdminModuleHandler) Create(c *gin.Context) {
	var req CreateModuleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.Error(err)
		return
	}

	courseID := domainCourse.CourseID(req.CourseID)

	module, err := h.uc.Create(c.Request.Context(), courseID, req.Title)
	if err != nil {
		c.Error(err)
		return
	}

	response := mapToModuleResponse(module)
	c.JSON(http.StatusCreated, response)
}

// Update handles PATCH /api/v1/admin/modules/:id
func (h *AdminModuleHandler) Update(c *gin.Context) {
	rawID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.Error(rootHttp.NewHTTPError(http.StatusBadRequest, "VALIDATION_ERROR", "Invalid module ID format: must be a valid UUID."))
		return
	}
	moduleID := domainModule.ModuleID(rawID)

	var req UpdateModuleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.Error(err)
		return
	}

	updatedModule, err := h.uc.UpdateMetadata(c.Request.Context(), moduleID, req.Title)
	if err != nil {
		c.Error(err)
		return
	}

	response := mapToModuleResponse(updatedModule)
	c.JSON(http.StatusOK, response)
}

// Reorder handles PUT /api/v1/admin/courses/:courseId/modules/order
func (h *AdminModuleHandler) Reorder(c *gin.Context) {
	rawCourseID, err := uuid.Parse(c.Param("courseId"))
	if err != nil {
		c.Error(rootHttp.NewHTTPError(http.StatusBadRequest, "VALIDATION_ERROR", "Invalid course ID format: must be a valid UUID."))
		return
	}
	courseID := domainCourse.CourseID(rawCourseID)

	var req ReorderModulesRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.Error(err)
		return
	}

	updatedModules, err := h.uc.Reorder(c.Request.Context(), courseID, req.ModuleIDs)
	if err != nil {
		c.Error(err)
		return
	}

	data := make([]ModuleResponse, len(updatedModules))
	for i, m := range updatedModules {
		data[i] = mapToModuleResponse(m)
	}

	c.JSON(http.StatusOK, gin.H{"modules": data})
}

// Delete handles DELETE /api/v1/admin/modules/:id
func (h *AdminModuleHandler) Delete(c *gin.Context) {
	rawID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.Error(rootHttp.NewHTTPError(http.StatusBadRequest, "VALIDATION_ERROR", "Invalid module ID format: must be a valid UUID."))
		return
	}
	moduleID := domainModule.ModuleID(rawID)

	if err := h.uc.Delete(c.Request.Context(), moduleID); err != nil {
		c.Error(err)
		return
	}

	c.Status(http.StatusNoContent)
}

func mapToModuleResponse(m *domainModule.Module) ModuleResponse {
	return ModuleResponse{
		ID:        uuid.UUID(m.ID()),
		CourseID:  uuid.UUID(m.CourseID()),
		Title:     m.Title(),
		Position:  int(m.Position()),
		CreatedAt: m.CreatedAt().Format(time.RFC3339),
		UpdatedAt: m.UpdatedAt().Format(time.RFC3339),
	}
}
