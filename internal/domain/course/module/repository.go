package module

import (
	"context"

	domainCourse "github.com/Jaran-Dev/backend-core/internal/domain/course"
)

type Repository interface {
	Create(ctx context.Context, m *Module) error
	GetByID(ctx context.Context, id ModuleID) (*Module, error)
	Update(ctx context.Context, m *Module) error
	Delete(ctx context.Context, id ModuleID) error
	ListByCourseID(ctx context.Context, courseID domainCourse.CourseID) ([]*Module, error)
	CountByCourseID(ctx context.Context, courseID domainCourse.CourseID) (int64, error)
	BulkReorder(ctx context.Context, courseID domainCourse.CourseID, positions map[ModuleID]Position) error
}
