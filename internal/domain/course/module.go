package course

import "github.com/google/uuid"

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
