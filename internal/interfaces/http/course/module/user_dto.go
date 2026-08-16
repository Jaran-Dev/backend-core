package modulehttp

import (
	"github.com/google/uuid"
)

type ListModulesQuery struct {
	// Locale is included for future compatibility.
	// If we add module_translations later, this query param will be used to resolve the title.
	Locale string `form:"locale" binding:"omitempty,locale"`
}

type PublicModuleResponse struct {
	ID       uuid.UUID `json:"id"`
	Title    string    `json:"title"`
	Position int       `json:"position"`
}

type PublicModuleListResponse struct {
	Modules []PublicModuleResponse `json:"modules"`
}
