package lesson

import (
	"context"

	"github.com/Jaran-Dev/backend-core/internal/domain/course/module"
)

type Repository interface {
	Create(ctx context.Context, l *Lesson) error
	GetByID(ctx context.Context, id LessonID) (*Lesson, error)
	GetByModuleID(ctx context.Context, moduleID module.ModuleID) ([]*Lesson, error)
	Update(ctx context.Context, l *Lesson) error
	Delete(ctx context.Context, id LessonID) error
}
