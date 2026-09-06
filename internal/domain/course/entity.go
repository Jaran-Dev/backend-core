package course

import (
	"errors"
	"time"

	"github.com/Jaran-Dev/backend-core/internal/domain/account"
	"github.com/Jaran-Dev/backend-core/internal/domain/shared"

	"github.com/google/uuid"
)

type Course struct {
	id           CourseID
	slug         Slug
	status       CourseStatus
	coverURL     string
	createdBy    account.AccountID
	createdAt    time.Time
	updatedAt    time.Time
	version      int // Optimistic concurrency control
	translations []CourseTranslation
	events       []interface{} // Domain events for downstream processing
}

func NewCourse(slug Slug, coverURL string, creatorID account.AccountID) *Course {
	now := time.Now()
	return &Course{
		id:        CourseID(uuid.New()),
		slug:      slug,
		status:    CourseStatusDraft,
		coverURL:  coverURL,
		createdBy: creatorID,
		createdAt: now,
		updatedAt: now,
		version:   1,
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

func (c *Course) CreatedBy() account.AccountID {
	return c.createdBy
}

func (c *Course) CreatedAt() time.Time {
	return c.createdAt
}

func (c *Course) UpdatedAt() time.Time {
	return c.updatedAt
}

func (c *Course) Version() int {
	return c.version
}

func (c *Course) Translations() []CourseTranslation {
	return c.translations
}

func (c *Course) PullEvents() []interface{} {
	events := c.events
	c.events = nil
	return events
}

func (c *Course) UpdateMetadata(slug Slug, coverURL string) {
	c.slug = slug
	c.coverURL = coverURL
	c.updatedAt = time.Now()
}

func (c *Course) AddOrUpdateTranslation(locale shared.Locale, title, description string) error {
	if !locale.IsValid() {
		return ErrInvalidLocale
	}

	for i, t := range c.translations {
		if t.Locale() == locale {
			c.translations[i].UpdateTitle(title)
			c.translations[i].UpdateDescription(description)
			c.updatedAt = time.Now()
			return nil
		}
	}

	newTrans := NewCourseTranslation(c.id, locale, title, description)
	c.translations = append(c.translations, newTrans)
	c.updatedAt = time.Now()
	return nil
}

func (c *Course) RemoveTranslation(locale shared.Locale) error {
	if !locale.IsValid() {
		return ErrInvalidLocale
	}

	for i, t := range c.translations {
		if t.Locale() == locale {
			c.translations = append(c.translations[:i], c.translations[i+1:]...)
			c.updatedAt = time.Now()
			return nil
		}
	}

	return ErrTranslationNotFound
}

func (c *Course) Publish() error {
	if c.status == CourseStatusArchived {
		return ErrInvalidStateTransition
	}
	if c.status == CourseStatusPublished {
		return nil
	}

	hasEnglish := false
	for _, t := range c.translations {
		if t.Locale() == shared.LocaleEnglish && t.Title() != "" {
			hasEnglish = true
			break
		}
	}
	if !hasEnglish {
		return ErrMissingEnglishTranslation
	}

	c.status = CourseStatusPublished
	c.updatedAt = time.Now()
	c.events = append(c.events, CoursePublishedEvent{CourseID: c.id})
	return nil
}

func (c *Course) Archive() error {
	if c.status == CourseStatusArchived {
		return nil
	}

	c.status = CourseStatusArchived
	c.updatedAt = time.Now()
	c.events = append(c.events, CourseArchivedEvent{CourseID: c.id})
	return nil
}

func UnmarshalCourse(
	id uuid.UUID,
	slugStr string,
	statusStr string,
	coverURL string,
	createdBy account.AccountID,
	createdAt time.Time,
	updatedAt time.Time,
	version int,
	translations []CourseTranslation,
) (*Course, error) {
	slug, err := NewSlug(slugStr)
	if err != nil {
		return nil, err
	}

	status := CourseStatus(statusStr)
	if !status.IsValid() {
		return nil, errors.New("invalid course status in database")
	}

	return &Course{
		id:           CourseID(id),
		slug:         slug,
		status:       status,
		coverURL:     coverURL,
		createdBy:    createdBy,
		createdAt:    createdAt,
		updatedAt:    updatedAt,
		version:      version,
		translations: translations,
		// events are not persisted, so they start empty upon reconstitution
	}, nil
}
