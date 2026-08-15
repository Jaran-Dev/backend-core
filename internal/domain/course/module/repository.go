package module

import (
	"context"

	"github.com/Jaran-Dev/backend-core/internal/domain/course"
)

type Repository interface {
	Create(ctx context.Context, m *Module) error
	GetByID(ctx context.Context, id ModuleID) (*Module, error)
	ListByCourseID(ctx context.Context, courseID course.CourseID) ([]*Module, error)
	Update(ctx context.Context, m *Module) error
	Delete(ctx context.Context, id ModuleID) error
	CountByCourseID(ctx context.Context, courseID course.CourseID) (int64, error)
}
