package module

import (
	"time"

	domainCourse "github.com/Jaran-Dev/backend-core/internal/domain/course"
	"github.com/google/uuid"
)

type Module struct {
	id        ModuleID
	courseID  domainCourse.CourseID
	title     string
	position  Position
	createdAt time.Time
	updatedAt time.Time
	version   int // Optimistic concurrency control
	events    []interface{}
}

func NewModule(courseID domainCourse.CourseID, title string, position int) (*Module, error) {
	pos, err := NewPosition(position)
	if err != nil {
		return nil, err
	}

	if title == "" {
		return nil, ErrInvalidModuleTitle
	}

	now := time.Now()
	return &Module{
		id:        ModuleID(uuid.New()),
		courseID:  courseID,
		title:     title,
		position:  pos,
		createdAt: now,
		updatedAt: now,
		version:   1,
	}, nil
}

func (m *Module) ID() ModuleID {
	return m.id
}

func (m *Module) CourseID() domainCourse.CourseID {
	return m.courseID
}

func (m *Module) Title() string {
	return m.title
}

func (m *Module) Position() Position {
	return m.position
}

func (m *Module) CreatedAt() time.Time {
	return m.createdAt
}

func (m *Module) UpdatedAt() time.Time {
	return m.updatedAt
}

func (m *Module) Version() int {
	return m.version
}

func (m *Module) PullEvents() []interface{} {
	events := m.events
	m.events = nil
	return events
}

func (m *Module) UpdateMetadata(title string) error {
	if title == "" {
		return ErrInvalidModuleTitle
	}
	m.title = title
	m.updatedAt = time.Now()
	return nil
}

func (m *Module) UpdatePosition(newPos int) error {
	pos, err := NewPosition(newPos)
	if err != nil {
		return err
	}

	if m.position == pos {
		return nil // Idempotent: no change needed
	}

	m.position = pos
	m.updatedAt = time.Now()
	m.events = append(m.events, ModuleReorderedEvent{
		ModuleID: m.id,
		NewPos:   pos,
	})
	return nil
}

type ModuleReorderedEvent struct {
	ModuleID ModuleID
	NewPos   Position
}

func UnmarshalModule(
	id uuid.UUID,
	courseID domainCourse.CourseID,
	title string,
	position int,
	createdAt time.Time,
	updatedAt time.Time,
	version int,
) (*Module, error) {
	pos, err := NewPosition(position)
	if err != nil {
		return nil, err
	}

	if title == "" {
		return nil, ErrInvalidModuleTitle
	}

	return &Module{
		id:        ModuleID(id),
		courseID:  courseID,
		title:     title,
		position:  pos,
		createdAt: createdAt,
		updatedAt: updatedAt,
		version:   version,
	}, nil
}
