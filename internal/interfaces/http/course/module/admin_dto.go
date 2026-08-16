package modulehttp

import (
	"github.com/google/uuid"
)

type CreateModuleRequest struct {
	CourseID uuid.UUID `json:"course_id" binding:"required,uuid4"`
	Title    string    `json:"title" binding:"required,min=1,max=200"`
}

type UpdateModuleRequest struct {
	Title *string `json:"title" binding:"omitempty,min=1,max=200"`
}

type ReorderModulesRequest struct {
	ModuleIDs []uuid.UUID `json:"module_ids" binding:"required,min=1,dive,required,uuid4"`
}

type ModuleResponse struct {
	ID        uuid.UUID `json:"id"`
	CourseID  uuid.UUID `json:"course_id"`
	Title     string    `json:"title"`
	Position  int       `json:"position"`
	CreatedAt string    `json:"created_at"`
	UpdatedAt string    `json:"updated_at"`
}
