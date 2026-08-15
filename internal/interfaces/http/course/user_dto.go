package coursehttp

import (
	"math"

	"github.com/google/uuid"
)

// ==========================================
// 1. GENERIC PAGINATION DTOS
// ==========================================

type PaginationMeta struct {
	Total      int64 `json:"total"`       // Total number of records in the database matching the query
	Page       int   `json:"page"`        // The current page number (1-based)
	PageSize   int   `json:"page_size"`   // The number of items requested per page
	TotalPages int   `json:"total_pages"` // Calculated total number of pages
	HasNext    bool  `json:"has_next"`    // Convenience boolean for the "Next" button
	HasPrev    bool  `json:"has_prev"`    // Convenience boolean for the "Previous" button
}

type PaginatedResponse[T any] struct {
	Data []T            `json:"data"`
	Meta PaginationMeta `json:"meta"`
}

func NewPaginatedResponse[T any](data []T, total int64, page, pageSize int) PaginatedResponse[T] {
	if pageSize <= 0 {
		pageSize = 10
	}
	if page <= 0 {
		page = 1
	}

	totalPages := int(math.Ceil(float64(total) / float64(pageSize)))

	return PaginatedResponse[T]{
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

// ==========================================
// 2. STUDENT RESPONSE DTOS
// ==========================================

type PublicCourseListItemDTO struct {
	ID       uuid.UUID `json:"id"`
	Slug     string    `json:"slug"`
	CoverURL string    `json:"cover_url"`
	Title    string    `json:"title"`
}

type PublicCourseDetailDTO struct {
	ID             uuid.UUID `json:"id"`
	Slug           string    `json:"slug"`
	CoverURL       string    `json:"cover_url"`
	Title          string    `json:"title"`
	Description    string    `json:"description"`
	ResolvedLocale string    `json:"resolved_locale"` // e.g., "en" (if ckb was missing)
}

type CourseReadModel struct {
	ID             uuid.UUID
	Slug           string
	CoverURL       string
	Title          string
	Description    string
	ResolvedLocale string
}

// ==========================================
// 3. STUDENT REQUEST DTOS (Query Parameters)
// ==========================================

// GET /api/v1/courses?page=1&page_size=10&locale=ckb&q=go
type ListCoursesQuery struct {
	Page     int    `form:"page" binding:"omitempty,min=1"`
	PageSize int    `form:"page_size" binding:"omitempty,min=1,max=50"`
	Locale   string `form:"locale" binding:"omitempty,locale"` // Custom validator
	Search   string `form:"q" binding:"omitempty,max=100"`     // Search keyword
}

// GET /api/v1/courses/:slug?locale=ckb
type GetCourseBySlugQuery struct {
	Locale string `form:"locale" binding:"omitempty,locale"`
}

func (q *ListCoursesQuery) SanitizeAndDefault() {
	if q.Page <= 0 {
		q.Page = 1
	}
	if q.PageSize <= 0 || q.PageSize > 50 {
		q.PageSize = 10
	}
	if q.Locale == "" {
		q.Locale = "en"
	}
}
