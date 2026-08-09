package lesson

import (
	"context"

	"github.com/Jaran-Dev/backend-core/internal/domain/course"
)

type Repository interface {
	Create(ctx context.Context, l *Lesson) error
	GetByID(ctx context.Context, id course.LessonID) (*Lesson, error)
	GetByModuleID(ctx context.Context, moduleID course.ModuleID) ([]*Lesson, error)
	Update(ctx context.Context, l *Lesson) error
	Delete(ctx context.Context, id course.LessonID) error
}
