package course

import (
	"time"

	domainCourse "github.com/Jaran-Dev/backend-core/internal/domain/course"
	domainAccount "github.com/Jaran-Dev/backend-core/internal/domain/account"

	"github.com/google/uuid"
)

type CourseORM struct {
	ID           uuid.UUID              `gorm:"type:uuid;primaryKey"`
	Slug         string                 `gorm:"uniqueIndex"`
	Status       string                 `gorm:"type:varchar(20);index"`
	CoverURL     *string                `gorm:"type:text"`
	CreatedBy    uuid.UUID              `gorm:"type:uuid;column:created_by"`
	CreatedAt    time.Time              `gorm:"autoCreateTime"`
	UpdatedAt    time.Time              `gorm:"autoUpdateTime"`
	Version      int                    `gorm:"default:1"`
	Translations []CourseTranslationORM `gorm:"foreignKey:CourseID;constraint:OnDelete:CASCADE"`
}

func (CourseORM) TableName() string {
	return "courses"
}

type CourseTranslationORM struct {
	CourseID    uuid.UUID `gorm:"type:uuid;primaryKey"`
	Locale      string    `gorm:"type:varchar(5);primaryKey"`
	Title       string    `gorm:"type:text"`
	Description *string   `gorm:"type:text"`
}

func (CourseTranslationORM) TableName() string { return "course_translations" }

func toPersistence(c *domainCourse.Course) *CourseORM {
	var coverURL *string
	if c.CoverURL() != "" {
		s := c.CoverURL()
		coverURL = &s
	}

	orm := &CourseORM{
		ID:        uuid.UUID(c.ID()),
		Slug:      c.Slug().String(),
		Status:    string(c.Status()),
		CoverURL:  coverURL,
		CreatedBy: uuid.UUID(c.CreatedBy()),
		CreatedAt: c.CreatedAt(),
		UpdatedAt: c.UpdatedAt(),
		Version:   c.Version(),
	}

	for _, t := range c.Translations() {
		var desc *string
		if t.Description() != "" {
			d := t.Description()
			desc = &d
		}
		orm.Translations = append(orm.Translations, CourseTranslationORM{
			CourseID:    uuid.UUID(t.CourseID()),
			Locale:      string(t.Locale()),
			Title:       t.Title(),
			Description: desc,
		})
	}

	return orm
}

func toDomain(orm *CourseORM) (*domainCourse.Course, error) {
	coverStr := ""
	if orm.CoverURL != nil {
		coverStr = *orm.CoverURL
	}

	var translations []domainCourse.CourseTranslation
	for _, tORM := range orm.Translations {
		descStr := ""
		if tORM.Description != nil {
			descStr = *tORM.Description
		}

		t, err := domainCourse.UnmarshalCourseTranslation(
			domainCourse.CourseID(tORM.CourseID),
			domainCourse.Locale(tORM.Locale),
			tORM.Title,
			descStr,
		)
		if err != nil {
			return nil, err
		}
		translations = append(translations, t)
	}

	return domainCourse.UnmarshalCourse(
		orm.ID,
		orm.Slug,
		orm.Status,
		coverStr,
		domainAccount.AccountID(orm.CreatedBy),
		orm.CreatedAt,
		orm.UpdatedAt,
		orm.Version,
		translations,
	)
}
