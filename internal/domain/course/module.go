package course

import (
	"context"
	"errors"

	"github.com/google/uuid"
)

type Module struct {
	id       ModuleID
	courseID CourseID
	position int
	title    string // Fallback title, canonical titles live in LessonTranslations
}

func NewModule(courseID CourseID, position int, title string) *Module {
	return &Module{
		id:       ModuleID(uuid.New()),
		courseID: courseID,
		position: position,
		title:    title,
	}
}

func (m *Module) ID() ModuleID {
	return m.id
}

func (m *Module) CourseID() CourseID {
	return m.courseID
}

func (m *Module) Position() int {
	return m.position
}

func (m *Module) Title() string {
	return m.title
}

func (m *Module) UpdateTitle(title string) {
	if title != "" {
		m.title = title
	}
}

func (m *Module) UpdatePosition(newPos int) {
	m.position = newPos
}

func UnmarshalModule(id uuid.UUID, courseID CourseID, position int, title string) (*Module, error) {
	if position < 0 {
		return nil, errors.New("cannot unmarshal module: position cannot be negative")
	}

	return &Module{
		id:       ModuleID(id),
		courseID: courseID,
		position: position,
		title:    title,
	}, nil
}

type ModuleRepository interface {
	Create(ctx context.Context, m *Module) error
	GetByID(ctx context.Context, id ModuleID) (*Module, error)
	ListByCourseID(ctx context.Context, courseID CourseID) ([]*Module, error)
	Update(ctx context.Context, m *Module) error
	Delete(ctx context.Context, id ModuleID) error
}
