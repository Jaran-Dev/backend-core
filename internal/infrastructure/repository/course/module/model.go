package module

import (
	"time"

	domainCourse "github.com/Jaran-Dev/backend-core/internal/domain/course"
	domainModule "github.com/Jaran-Dev/backend-core/internal/domain/course/module"

	"github.com/google/uuid"
)

type ModuleORM struct {
	ID        uuid.UUID `gorm:"type:uuid;primaryKey"`
	CourseID  uuid.UUID `gorm:"type:uuid;index;column:course_id"`
	Title     string    `gorm:"type:text"`
	Position  int       `gorm:"uniqueIndex:idx_course_position"`
	CreatedAt time.Time `gorm:"autoCreateTime"`
	UpdatedAt time.Time `gorm:"autoUpdateTime"`
	Version   int       `gorm:"default:1"`
}

func (ModuleORM) TableName() string {
	return "modules"
}

func toPersistence(m *domainModule.Module) *ModuleORM {
	return &ModuleORM{
		ID:        uuid.UUID(m.ID()),
		CourseID:  uuid.UUID(m.CourseID()),
		Title:     m.Title(),
		Position:  int(m.Position()),
		CreatedAt: m.CreatedAt(),
		UpdatedAt: m.UpdatedAt(),
		Version:   m.Version(),
	}
}

func toDomain(orm *ModuleORM) (*domainModule.Module, error) {
	return domainModule.UnmarshalModule(
		orm.ID,
		domainCourse.CourseID(orm.CourseID),
		orm.Title,
		orm.Position,
		orm.CreatedAt,
		orm.UpdatedAt,
		orm.Version,
	)
}
