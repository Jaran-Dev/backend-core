package category

import (
	"errors"
	"time"

	"github.com/Jaran-Dev/backend-core/internal/domain/shared"
	"github.com/google/uuid"
)

type Category struct {
	id           CategoryID
	parentID     *CategoryID
	slug         Slug
	status       CategoryStatus
	coverURL     string
	position     Position
	deletedAt    *time.Time
	createdAt    time.Time
	updatedAt    time.Time
	version      int
	translations []CategoryTranslation
	events       []interface{}
}

func NewCategory(slug Slug, coverURL string, parentID *CategoryID) *Category {
	now := time.Now()
	return &Category{
		id:        CategoryID(uuid.New()),
		parentID:  parentID,
		slug:      slug,
		status:    CategoryStatusDraft,
		coverURL:  coverURL,
		position:  0,
		createdAt: now,
		updatedAt: now,
		version:   1,
	}
}

func (c *Category) ID() CategoryID                      { return c.id }
func (c *Category) ParentID() *CategoryID               { return c.parentID }
func (c *Category) Slug() Slug                          { return c.slug }
func (c *Category) Status() CategoryStatus              { return c.status }
func (c *Category) CoverURL() string                    { return c.coverURL }
func (c *Category) Position() Position                  { return c.position }
func (c *Category) DeletedAt() *time.Time               { return c.deletedAt }
func (c *Category) CreatedAt() time.Time                { return c.createdAt }
func (c *Category) UpdatedAt() time.Time                { return c.updatedAt }
func (c *Category) Version() int                        { return c.version }
func (c *Category) Translations() []CategoryTranslation { return c.translations }
func (c *Category) PullEvents() []interface{} {
	events := c.events
	c.events = nil
	return events
}

func (c *Category) UpdateMetadata(slug Slug, coverURL string) {
	c.slug = slug
	c.coverURL = coverURL
	c.updatedAt = time.Now()
}

func (c *Category) UpdatePosition(newPos int) error {
	pos, err := NewPosition(newPos)
	if err != nil {
		return err
	}
	c.position = pos
	c.updatedAt = time.Now()
	return nil
}

func (c *Category) Reparent(newParentID *CategoryID) {
	c.parentID = newParentID
	c.updatedAt = time.Now()
	c.events = append(c.events, CategoryReparentedEvent{
		CategoryID:  c.id,
		NewParentID: newParentID,
	})
}

func (c *Category) AddOrUpdateTranslation(locale shared.Locale, title, description string) error {
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

	newTrans := NewCategoryTranslation(c.id, locale, title, description)
	c.translations = append(c.translations, newTrans)
	c.updatedAt = time.Now()
	return nil
}

func (c *Category) RemoveTranslation(locale shared.Locale) error {
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

func (c *Category) Publish() error {
	if c.status == CategoryStatusArchived {
		return ErrInvalidStateTransition
	}
	if c.status == CategoryStatusPublished {
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

	c.status = CategoryStatusPublished
	c.updatedAt = time.Now()
	c.events = append(c.events, CategoryPublishedEvent{CategoryID: c.id})
	return nil
}

func (c *Category) Archive() error {
	if c.status == CategoryStatusArchived {
		return nil
	}

	c.status = CategoryStatusArchived
	c.updatedAt = time.Now()
	c.events = append(c.events, CategoryArchivedEvent{CategoryID: c.id})
	return nil
}

func (c *Category) SoftDelete() error {
	if c.status == CategoryStatusPublished {
		return ErrInvalidStateTransition
	}
	now := time.Now()
	c.deletedAt = &now
	c.updatedAt = now
	return nil
}

// Events
type CategoryPublishedEvent struct {
	CategoryID CategoryID
}

type CategoryArchivedEvent struct {
	CategoryID CategoryID
}

type CategoryReparentedEvent struct {
	CategoryID  CategoryID
	NewParentID *CategoryID
}

func UnmarshalCategory(
	id uuid.UUID,
	parentID *uuid.UUID,
	slugStr string,
	statusStr string,
	coverURL string,
	position int,
	deletedAt *time.Time,
	createdAt time.Time,
	updatedAt time.Time,
	version int,
	translations []CategoryTranslation,
) (*Category, error) {
	slug, err := NewSlug(slugStr)
	if err != nil {
		return nil, err
	}

	status := CategoryStatus(statusStr)
	if !status.IsValid() {
		return nil, errors.New("invalid category status in database")
	}

	var pID *CategoryID
	if parentID != nil {
		p := CategoryID(*parentID)
		pID = &p
	}

	return &Category{
		id:           CategoryID(id),
		parentID:     pID,
		slug:         slug,
		status:       status,
		coverURL:     coverURL,
		position:     Position(position),
		deletedAt:    deletedAt,
		createdAt:    createdAt,
		updatedAt:    updatedAt,
		version:      version,
		translations: translations,
	}, nil
}
