package categoryhttp

import (
	"github.com/google/uuid"
)

// ==========================================
// 1. ADMIN REQUEST DTOS
// ==========================================

type CreateCategoryRequest struct {
	Slug     string     `json:"slug" binding:"required,min=3,max=100"`
	CoverURL string     `json:"cover_url" binding:"omitempty,url"`
	ParentID *uuid.UUID `json:"parent_id" binding:"omitempty,uuid4"`
}

type UpdateCategoryRequest struct {
	Slug     *string `json:"slug" binding:"omitempty,min=3,max=100"`
	CoverURL *string `json:"cover_url" binding:"omitempty,url"`
}

type ReparentCategoryRequest struct {
	ParentID *uuid.UUID `json:"parent_id" binding:"omitempty,uuid4"`
}

type UpsertCategoryTranslationRequest struct {
	Locale      string `json:"locale" binding:"required,locale"`
	Title       string `json:"title" binding:"required,min=1,max=200"`
	Description string `json:"description" binding:"omitempty,max=2000"`
}

type UpdateCategoryStatusRequest struct {
	Action string `json:"action" binding:"required,oneof=publish archive"`
}

// ==========================================
// 2. ADMIN RESPONSE DTOS
// ==========================================

type CategoryTranslationResponse struct {
	Locale      string `json:"locale"`
	Title       string `json:"title"`
	Description string `json:"description,omitempty"`
}

type CategoryResponse struct {
	ID           uuid.UUID                     `json:"id"`
	ParentID     *uuid.UUID                    `json:"parent_id,omitempty"`
	Slug         string                        `json:"slug"`
	Status       string                        `json:"status"`
	CoverURL     string                        `json:"cover_url"`
	Position     int                           `json:"position"`
	CreatedAt    string                        `json:"created_at"`
	UpdatedAt    string                        `json:"updated_at"`
	Translations []CategoryTranslationResponse `json:"translations,omitempty"`
}
