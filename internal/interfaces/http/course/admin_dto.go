package coursehttp

import (
	"github.com/google/uuid"
)

// ==========================================
// 1. ADMIN REQUEST DTOS
// ==========================================

type CreateCourseRequest struct {
	Slug     string `json:"slug" binding:"required,min=3,max=100"`
	CoverURL string `json:"cover_url" binding:"omitempty,url"`
}

type UpdateCourseRequest struct {
	Slug     *string `json:"slug" binding:"omitempty,min=3,max=100"`
	CoverURL *string `json:"cover_url" binding:"omitempty,url"`
}

type UpsertTranslationRequest struct {
	Locale      string `json:"locale" binding:"required,locale"` // Custom validator registered in Gin
	Title       string `json:"title" binding:"required,min=1,max=200"`
	Description string `json:"description" binding:"omitempty,max=2000"`
}

type UpdateStatusRequest struct {
	Action string `json:"action" binding:"required,oneof=publish archive"`
}

// ==========================================
// 2. ADMIN RESPONSE DTOS
// ==========================================

type TranslationResponse struct {
	Locale      string `json:"locale"`
	Title       string `json:"title"`
	Description string `json:"description,omitempty"`
}

type CourseResponse struct {
	ID           uuid.UUID             `json:"id"`
	Slug         string                `json:"slug"`
	Status       string                `json:"status"`
	CoverURL     string                `json:"cover_url"`
	CreatedBy    uuid.UUID             `json:"created_by"`
	CreatedAt    string                `json:"created_at"` // Formatted as RFC3339 string for JS
	UpdatedAt    string                `json:"updated_at"`
	Translations []TranslationResponse `json:"translations,omitempty"`
}
