package course

import (
	"time"

	domainAccount "github.com/Jaran-Dev/backend-core/internal/domain/account"
	domainCourse "github.com/Jaran-Dev/backend-core/internal/domain/course"
	"github.com/Jaran-Dev/backend-core/internal/domain/shared"

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
			shared.Locale(tORM.Locale),
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

type CourseFlatORM struct {
	ID        uuid.UUID `gorm:"column:id"`
	Slug      string    `gorm:"column:slug"`
	Status    string    `gorm:"column:status"`
	CoverURL  *string   `gorm:"column:cover_url"`
	CreatedBy uuid.UUID `gorm:"column:created_by"`
	CreatedAt time.Time `gorm:"column:created_at"`
	UpdatedAt time.Time `gorm:"column:updated_at"`
	Version   int       `gorm:"column:version"`

	// Flattened translation columns (aliased to avoid collision with c.id, c.slug, etc.)
	TransLocale      *string `gorm:"column:trans_locale"`
	TransTitle       *string `gorm:"column:trans_title"`
	TransDescription *string `gorm:"column:trans_description"`
}

func flatToCourseORM(flat *CourseFlatORM) *CourseORM {
	orm := &CourseORM{
		ID:        flat.ID,
		Slug:      flat.Slug,
		Status:    flat.Status,
		CoverURL:  flat.CoverURL,
		CreatedBy: flat.CreatedBy,
		CreatedAt: flat.CreatedAt,
		UpdatedAt: flat.UpdatedAt,
		Version:   flat.Version,
	}

	// If the lateral join found a translation, add it to the translations slice
	if flat.TransLocale != nil && flat.TransTitle != nil {
		orm.Translations = []CourseTranslationORM{{
			CourseID:    flat.ID,
			Locale:      *flat.TransLocale,
			Title:       *flat.TransTitle,
			Description: flat.TransDescription,
		}}
	}

	return orm
}
