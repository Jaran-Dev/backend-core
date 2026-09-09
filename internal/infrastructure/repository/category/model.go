package category

import (
	"time"

	domainCategory "github.com/Jaran-Dev/backend-core/internal/domain/category"
	"github.com/Jaran-Dev/backend-core/internal/domain/shared"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type CategoryORM struct {
	ID           uuid.UUID  `gorm:"type:uuid;primaryKey"`
	ParentID     *uuid.UUID `gorm:"type:uuid;column:parent_id;index"`
	Slug         string     `gorm:"uniqueIndex"`
	Status       string     `gorm:"type:varchar(20);index"`
	CoverURL     *string    `gorm:"type:text"`
	Position     int
	DeletedAt    gorm.DeletedAt           `gorm:"column:deleted_at;index"`
	CreatedAt    time.Time                `gorm:"autoCreateTime"`
	UpdatedAt    time.Time                `gorm:"autoUpdateTime"`
	Version      int                      `gorm:"default:1"`
	Translations []CategoryTranslationORM `gorm:"foreignKey:CategoryID;constraint:OnDelete:CASCADE"`
}

func (CategoryORM) TableName() string { return "categories" }

type CategoryTranslationORM struct {
	CategoryID  uuid.UUID `gorm:"type:uuid;primaryKey"`
	Locale      string    `gorm:"type:varchar(5);primaryKey"`
	Title       string    `gorm:"type:text"`
	Description *string   `gorm:"type:text"`
}

func (CategoryTranslationORM) TableName() string { return "category_translations" }

type CategoryClosureORM struct {
	AncestorID   uuid.UUID `gorm:"type:uuid;primaryKey"`
	DescendantID uuid.UUID `gorm:"type:uuid;primaryKey"`
	Depth        int
}

func (CategoryClosureORM) TableName() string { return "category_closure" }

type CategoryFlatORM struct {
	ID        uuid.UUID  `gorm:"column:id"`
	ParentID  *uuid.UUID `gorm:"column:parent_id"`
	Slug      string     `gorm:"column:slug"`
	Status    string     `gorm:"column:status"`
	CoverURL  *string    `gorm:"column:cover_url"`
	Position  int        `gorm:"column:position"`
	CreatedAt time.Time  `gorm:"column:created_at"`
	UpdatedAt time.Time  `gorm:"column:updated_at"`
	Version   int        `gorm:"column:version"`

	TransLocale      *string `gorm:"column:trans_locale"`
	TransTitle       *string `gorm:"column:trans_title"`
	TransDescription *string `gorm:"column:trans_description"`
}

func flatToCategoryORM(flat *CategoryFlatORM) *CategoryORM {
	orm := &CategoryORM{
		ID:        flat.ID,
		ParentID:  flat.ParentID,
		Slug:      flat.Slug,
		Status:    flat.Status,
		CoverURL:  flat.CoverURL,
		Position:  flat.Position,
		CreatedAt: flat.CreatedAt,
		UpdatedAt: flat.UpdatedAt,
		Version:   flat.Version,
	}

	if flat.TransLocale != nil && flat.TransTitle != nil {
		orm.Translations = []CategoryTranslationORM{{
			CategoryID:  flat.ID,
			Locale:      *flat.TransLocale,
			Title:       *flat.TransTitle,
			Description: flat.TransDescription,
		}}
	}

	return orm
}

func toPersistence(c *domainCategory.Category) *CategoryORM {
	var parentID *uuid.UUID
	if c.ParentID() != nil {
		pid := uuid.UUID(*c.ParentID())
		parentID = &pid
	}

	var coverURL *string
	if c.CoverURL() != "" {
		s := c.CoverURL()
		coverURL = &s
	}

	orm := &CategoryORM{
		ID:        uuid.UUID(c.ID()),
		ParentID:  parentID,
		Slug:      c.Slug().String(),
		Status:    string(c.Status()),
		CoverURL:  coverURL,
		Position:  int(c.Position()),
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
		orm.Translations = append(orm.Translations, CategoryTranslationORM{
			CategoryID:  uuid.UUID(t.CategoryID()),
			Locale:      string(t.Locale()),
			Title:       t.Title(),
			Description: desc,
		})
	}

	return orm
}

func toDomain(orm *CategoryORM) (*domainCategory.Category, error) {
	coverStr := ""
	if orm.CoverURL != nil {
		coverStr = *orm.CoverURL
	}

	var deletedAt *time.Time
	if orm.DeletedAt.Valid {
		t := orm.DeletedAt.Time
		deletedAt = &t
	}

	var translations []domainCategory.CategoryTranslation
	for _, tORM := range orm.Translations {
		descStr := ""
		if tORM.Description != nil {
			descStr = *tORM.Description
		}

		t, err := domainCategory.UnmarshalCategoryTranslation(
			domainCategory.CategoryID(tORM.CategoryID),
			shared.Locale(tORM.Locale),
			tORM.Title,
			descStr,
		)
		if err != nil {
			return nil, err
		}
		translations = append(translations, t)
	}

	return domainCategory.UnmarshalCategory(
		orm.ID,
		orm.ParentID,
		orm.Slug,
		orm.Status,
		coverStr,
		orm.Position,
		deletedAt,
		orm.CreatedAt,
		orm.UpdatedAt,
		orm.Version,
		translations,
	)
}
