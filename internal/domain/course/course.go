package course

import (
	"time"

	"github.com/Jaran-Dev/backend-core/internal/domain/admin"

	"github.com/google/uuid"
)

type Course struct {
	id        CourseID
	slug      Slug
	status    CourseStatus
	coverURL  string
	createdBy admin.AdminID
	createdAt time.Time
	updatedAt time.Time

	translations []CourseTranslation
	modules      []Module
}

func NewCourse(slug Slug, coverURL string, creatorID admin.AdminID) *Course {
	now := time.Now()
	return &Course{
		id:        CourseID(uuid.New()),
		slug:      slug,
		status:    CourseStatusDraft,
		coverURL:  coverURL,
		createdBy: creatorID,
		createdAt: now,
		updatedAt: now,
	}
}

func (c *Course) ID() CourseID {
	return c.id
}

func (c *Course) Slug() Slug {
	return c.slug
}

func (c *Course) Status() CourseStatus {
	return c.status
}

func (c *Course) CoverURL() string {
	return c.coverURL
}

func (c *Course) CreatedBy() admin.AdminID {
	return c.createdBy
}

func (c *Course) CreatedAt() time.Time {
	return c.createdAt
}

func (c *Course) UpdatedAt() time.Time {
	return c.updatedAt
}

func (c *Course) Translations() []CourseTranslation {
	return c.translations
}

func (c *Course) Modules() []Module {
	return c.modules
}

func (c *Course) AddOrUpdateTranslation(locale Locale, title, desc string) error {
	if !locale.IsValid() {
		return ErrInvalidLocale
	}

	for i, t := range c.translations {
		if t.locale == locale {
			c.translations[i].title = title
			c.translations[i].description = desc
			c.updatedAt = time.Now()
			return nil
		}
	}
	c.translations = append(c.translations, CourseTranslation{
		courseID:    c.id,
		locale:      locale,
		title:       title,
		description: desc,
	})
	c.updatedAt = time.Now()
	return nil
}

func (c *Course) AddModule(m *Module) {
	c.modules = append(c.modules, *m)
	c.updatedAt = time.Now()
}

func (c *Course) Publish() error {
	if c.status == CourseStatusArchived {
		return ErrCourseNotFound
	}

	hasEnglish := false
	for _, t := range c.translations {
		if t.locale == LocaleEnglish && t.title != "" {
			hasEnglish = true
			break
		}
	}
	if !hasEnglish {
		return ErrMissingEnglishTranslation
	}
	if len(c.modules) == 0 {
		return ErrMissingModules
	}

	c.status = CourseStatusPublished
	c.updatedAt = time.Now()
	return nil
}

func (c *Course) Archive() {
	c.status = CourseStatusArchived
	c.updatedAt = time.Now()
}
