package categoryhttp

import (
	"math"

	"github.com/google/uuid"
)

// ==========================================
// 1. CATEGORY TREE DTOS
// ==========================================

type CategoryTreeNode struct {
	ID       uuid.UUID          `json:"id"`
	Slug     string             `json:"slug"`
	CoverURL string             `json:"cover_url"`
	Title    string             `json:"title"`
	Children []CategoryTreeNode `json:"children,omitempty"`
}

type CategoryTreeResponse struct {
	Categories []CategoryTreeNode `json:"categories"`
}

// ==========================================
// 2. CATEGORY COURSES LIST DTOS
// ==========================================

type ListCoursesByCategoryQuery struct {
	Page     int    `form:"page" binding:"omitempty,min=1"`
	PageSize int    `form:"page_size" binding:"omitempty,min=1,max=50"`
	Locale   string `form:"locale" binding:"omitempty,locale"`
}

type CategoryCourseItem struct {
	ID       uuid.UUID `json:"id"`
	Slug     string    `json:"slug"`
	CoverURL string    `json:"cover_url"`
	Title    string    `json:"title"`
}

type CategoryCoursesResponse struct {
	Data []CategoryCourseItem `json:"data"`
	Meta PaginationMeta       `json:"meta"`
}

type PaginationMeta struct {
	Total      int64 `json:"total"`
	Page       int   `json:"page"`
	PageSize   int   `json:"page_size"`
	TotalPages int   `json:"total_pages"`
	HasNext    bool  `json:"has_next"`
	HasPrev    bool  `json:"has_prev"`
}

func NewCategoryCoursesResponse(data []CategoryCourseItem, total int64, page, pageSize int) CategoryCoursesResponse {
	if pageSize <= 0 {
		pageSize = 10
	}
	if page <= 0 {
		page = 1
	}

	totalPages := int(math.Ceil(float64(total) / float64(pageSize)))

	return CategoryCoursesResponse{
		Data: data,
		Meta: PaginationMeta{
			Total:      total,
			Page:       page,
			PageSize:   pageSize,
			TotalPages: totalPages,
			HasNext:    page < totalPages,
			HasPrev:    page > 1,
		},
	}
}
