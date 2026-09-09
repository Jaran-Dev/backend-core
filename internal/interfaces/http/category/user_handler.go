package categoryhttp

import (
	"context"
	"net/http"

	domainCategory "github.com/Jaran-Dev/backend-core/internal/domain/category"
	domainCourse "github.com/Jaran-Dev/backend-core/internal/domain/course"
	"github.com/Jaran-Dev/backend-core/internal/domain/shared"
	"github.com/google/uuid"

	"github.com/gin-gonic/gin"
)

type StudentCategoryUseCase interface {
	GetTree(ctx context.Context, preferredLocale shared.Locale) ([]*domainCategory.CategoryNode, error)
	ListCoursesByCategory(ctx context.Context, slug string, preferredLocale shared.Locale, page, pageSize int) ([]*domainCourse.Course, int64, error)
}

type StudentCategoryHandler struct {
	uc StudentCategoryUseCase
}

func NewStudentCategoryHandler(uc StudentCategoryUseCase) *StudentCategoryHandler {
	return &StudentCategoryHandler{uc: uc}
}

func (h *StudentCategoryHandler) RegisterRoutes(rg *gin.RouterGroup) {
	categories := rg.Group("/categories")
	{
		categories.GET("", h.GetTree)
		categories.GET("/:slug/courses", h.ListCoursesByCategory)
	}
}

// GetTree handles GET /api/v1/categories?locale=ckb
func (h *StudentCategoryHandler) GetTree(c *gin.Context) {
	locale := c.Query("locale")
	preferredLocale := shared.LocaleEnglish
	if locale != "" {
		preferredLocale = shared.Locale(locale)
	}

	tree, err := h.uc.GetTree(c.Request.Context(), preferredLocale)
	if err != nil {
		c.Error(err)
		return
	}

	nodes := make([]CategoryTreeNode, len(tree))
	for i, node := range tree {
		nodes[i] = mapToTreeNode(node)
	}

	response := CategoryTreeResponse{Categories: nodes}
	c.JSON(http.StatusOK, response)
}

// ListCoursesByCategory handles GET /api/v1/categories/:slug/courses?locale=ckb&page=1&page_size=10
func (h *StudentCategoryHandler) ListCoursesByCategory(c *gin.Context) {
	slugStr := c.Param("slug")

	var query ListCoursesByCategoryQuery
	if err := c.ShouldBindQuery(&query); err != nil {
		c.Error(err)
		return
	}

	if query.Page <= 0 {
		query.Page = 1
	}
	if query.PageSize <= 0 || query.PageSize > 50 {
		query.PageSize = 10
	}

	preferredLocale := shared.LocaleEnglish
	if query.Locale != "" {
		preferredLocale = shared.Locale(query.Locale)
	}

	courses, total, err := h.uc.ListCoursesByCategory(c.Request.Context(), slugStr, preferredLocale, query.Page, query.PageSize)
	if err != nil {
		c.Error(err)
		return
	}

	data := make([]CategoryCourseItem, len(courses))
	for i, course := range courses {
		title := ""
		if len(course.Translations()) > 0 {
			title = course.Translations()[0].Title()
		}

		data[i] = CategoryCourseItem{
			ID:       uuid.UUID(course.ID()),
			Slug:     course.Slug().String(),
			CoverURL: course.CoverURL(),
			Title:    title,
		}
	}

	response := NewCategoryCoursesResponse(data, total, query.Page, query.PageSize)
	c.JSON(http.StatusOK, response)
}

func mapToTreeNode(node *domainCategory.CategoryNode) CategoryTreeNode {
	treeNode := CategoryTreeNode{
		ID:       uuid.UUID(node.Category.ID()),
		Slug:     node.Category.Slug().String(),
		CoverURL: node.Category.CoverURL(),
		Title:    "",
	}

	if len(node.Category.Translations()) > 0 {
		treeNode.Title = node.Category.Translations()[0].Title()
	}

	if len(node.Children) > 0 {
		treeNode.Children = make([]CategoryTreeNode, len(node.Children))
		for i, child := range node.Children {
			treeNode.Children[i] = mapToTreeNode(child)
		}
	}

	return treeNode
}
