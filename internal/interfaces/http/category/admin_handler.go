package categoryhttp

import (
	"context"
	"net/http"
	"time"

	domainCategory "github.com/Jaran-Dev/backend-core/internal/domain/category"
	"github.com/Jaran-Dev/backend-core/internal/domain/shared"
	rootHttp "github.com/Jaran-Dev/backend-core/internal/interfaces/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type AdminCategoryUseCase interface {
	Create(ctx context.Context, slug, coverURL string, parentID *uuid.UUID) (*domainCategory.Category, error)
	UpdateMetadata(ctx context.Context, id domainCategory.CategoryID, slug *string, coverURL *string) (*domainCategory.Category, error)
	UpsertTranslation(ctx context.Context, id domainCategory.CategoryID, locale shared.Locale, title, description string) (*domainCategory.Category, error)
	Publish(ctx context.Context, id domainCategory.CategoryID) (*domainCategory.Category, error)
	Archive(ctx context.Context, id domainCategory.CategoryID) (*domainCategory.Category, error)
	SoftDelete(ctx context.Context, id domainCategory.CategoryID) error
	Reparent(ctx context.Context, id domainCategory.CategoryID, newParentID *uuid.UUID) (*domainCategory.Category, error)
}

type AdminCategoryHandler struct {
	uc AdminCategoryUseCase
}

func NewAdminCategoryHandler(uc AdminCategoryUseCase) *AdminCategoryHandler {
	return &AdminCategoryHandler{uc: uc}
}

func (h *AdminCategoryHandler) RegisterRoutes(rg *gin.RouterGroup) {
	categories := rg.Group("/categories")
	{
		categories.POST("", h.Create)
		categories.PATCH("/:id", h.Update)
		categories.POST("/:id/translations", h.UpsertTranslation)
		categories.PATCH("/:id/status", h.ChangeStatus)
		categories.DELETE("/:id", h.SoftDelete)
		categories.PATCH("/:id/parent", h.Reparent)
	}
}

// Create handles POST /api/v1/admin/categories
func (h *AdminCategoryHandler) Create(c *gin.Context) {
	var req CreateCategoryRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.Error(err)
		return
	}

	category, err := h.uc.Create(c.Request.Context(), req.Slug, req.CoverURL, req.ParentID)
	if err != nil {
		c.Error(err)
		return
	}

	response := mapToCategoryResponse(category)
	c.JSON(http.StatusCreated, response)
}

// Update handles PATCH /api/v1/admin/categories/:id
func (h *AdminCategoryHandler) Update(c *gin.Context) {
	rawID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.Error(rootHttp.NewHTTPError(http.StatusBadRequest, "VALIDATION_ERROR", "Invalid category ID format: must be a valid UUID."))
		return
	}
	categoryID := domainCategory.CategoryID(rawID)

	var req UpdateCategoryRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.Error(err)
		return
	}

	updatedCategory, err := h.uc.UpdateMetadata(c.Request.Context(), categoryID, req.Slug, req.CoverURL)
	if err != nil {
		c.Error(err)
		return
	}

	response := mapToCategoryResponse(updatedCategory)
	c.JSON(http.StatusOK, response)
}

// UpsertTranslation handles POST /api/v1/admin/categories/:id/translations
func (h *AdminCategoryHandler) UpsertTranslation(c *gin.Context) {
	rawID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.Error(rootHttp.NewHTTPError(http.StatusBadRequest, "VALIDATION_ERROR", "Invalid category ID format: must be a valid UUID."))
		return
	}
	categoryID := domainCategory.CategoryID(rawID)

	var req UpsertCategoryTranslationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.Error(err)
		return
	}

	locale := shared.Locale(req.Locale)

	updatedCategory, err := h.uc.UpsertTranslation(c.Request.Context(), categoryID, locale, req.Title, req.Description)
	if err != nil {
		c.Error(err)
		return
	}

	response := mapToCategoryResponse(updatedCategory)
	c.JSON(http.StatusOK, response)
}

// ChangeStatus handles PATCH /api/v1/admin/categories/:id/status
func (h *AdminCategoryHandler) ChangeStatus(c *gin.Context) {
	rawID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.Error(rootHttp.NewHTTPError(http.StatusBadRequest, "VALIDATION_ERROR", "Invalid category ID format: must be a valid UUID."))
		return
	}
	categoryID := domainCategory.CategoryID(rawID)

	var req UpdateCategoryStatusRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.Error(err)
		return
	}

	var updatedCategory *domainCategory.Category
	switch req.Action {
	case "publish":
		updatedCategory, err = h.uc.Publish(c.Request.Context(), categoryID)
	case "archive":
		updatedCategory, err = h.uc.Archive(c.Request.Context(), categoryID)
	}

	if err != nil {
		c.Error(err)
		return
	}

	response := mapToCategoryResponse(updatedCategory)
	c.JSON(http.StatusOK, response)
}

// SoftDelete handles DELETE /api/v1/admin/categories/:id
func (h *AdminCategoryHandler) SoftDelete(c *gin.Context) {
	rawID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.Error(rootHttp.NewHTTPError(http.StatusBadRequest, "VALIDATION_ERROR", "Invalid category ID format: must be a valid UUID."))
		return
	}
	categoryID := domainCategory.CategoryID(rawID)

	if err := h.uc.SoftDelete(c.Request.Context(), categoryID); err != nil {
		c.Error(err)
		return
	}

	c.Status(http.StatusNoContent)
}

// Reparent handles PATCH /api/v1/admin/categories/:id/parent
func (h *AdminCategoryHandler) Reparent(c *gin.Context) {
	rawID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.Error(rootHttp.NewHTTPError(http.StatusBadRequest, "VALIDATION_ERROR", "Invalid category ID format: must be a valid UUID."))
		return
	}
	categoryID := domainCategory.CategoryID(rawID)

	var req ReparentCategoryRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.Error(err)
		return
	}

	updatedCategory, err := h.uc.Reparent(c.Request.Context(), categoryID, req.ParentID)
	if err != nil {
		c.Error(err)
		return
	}

	response := mapToCategoryResponse(updatedCategory)
	c.JSON(http.StatusOK, response)
}

func mapToCategoryResponse(cat *domainCategory.Category) CategoryResponse {
	var transResp []CategoryTranslationResponse
	if cat.Translations() != nil {
		transResp = make([]CategoryTranslationResponse, len(cat.Translations()))
		for i, t := range cat.Translations() {
			transResp[i] = CategoryTranslationResponse{
				Locale:      string(t.Locale()),
				Title:       t.Title(),
				Description: t.Description(),
			}
		}
	}

	var parentID *uuid.UUID
	if cat.ParentID() != nil {
		pid := uuid.UUID(*cat.ParentID())
		parentID = &pid
	}

	return CategoryResponse{
		ID:           uuid.UUID(cat.ID()),
		ParentID:     parentID,
		Slug:         cat.Slug().String(),
		Status:       string(cat.Status()),
		CoverURL:     cat.CoverURL(),
		Position:     int(cat.Position()),
		CreatedAt:    cat.CreatedAt().Format(time.RFC3339),
		UpdatedAt:    cat.UpdatedAt().Format(time.RFC3339),
		Translations: transResp,
	}
}
