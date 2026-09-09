package category

import (
	"context"
	"errors"
	"fmt"

	domainCategory "github.com/Jaran-Dev/backend-core/internal/domain/category"
	"github.com/Jaran-Dev/backend-core/internal/domain/shared"
	"github.com/Jaran-Dev/backend-core/internal/infrastructure/repository"

	"github.com/google/uuid"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type pgRepository struct {
	db *gorm.DB
}

func NewCategoryRepository(db *gorm.DB) domainCategory.Repository {
	return &pgRepository{db: db}
}

func (r *pgRepository) Create(ctx context.Context, c *domainCategory.Category) error {
	orm := toPersistence(c)
	db := repository.GetDB(ctx, r.db)

	// The DB trigger (maintain_category_closure) will automatically insert
	// the self-reference and ancestor rows into category_closure.
	if err := db.Create(orm).Error; err != nil {
		translatedErr := repository.TranslateGormError(err)
		if errors.Is(translatedErr, gorm.ErrDuplicatedKey) {
			return fmt.Errorf("%w: slug must be unique", domainCategory.ErrInvalidSlug)
		}
		if errors.Is(translatedErr, repository.ErrForeignKeyViolated) {
			return domainCategory.ErrInvalidDepth
		}
		return fmt.Errorf("failed to create category in database: %w", translatedErr)
	}
	return nil
}

func (r *pgRepository) GetByID(ctx context.Context, id domainCategory.CategoryID, loadTranslations bool) (*domainCategory.Category, error) {
	db := repository.GetDB(ctx, r.db)

	if loadTranslations {
		db = db.Preload("Translations")
	}

	var orm CategoryORM
	err := db.First(&orm, "id = ?", uuid.UUID(id)).Error
	if err != nil {
		translatedErr := repository.TranslateGormError(err)
		if errors.Is(translatedErr, gorm.ErrRecordNotFound) {
			return nil, domainCategory.ErrCategoryNotFound
		}
		return nil, fmt.Errorf("failed to query category by id: %w", translatedErr)
	}

	return toDomain(&orm)
}

func (r *pgRepository) GetBySlug(ctx context.Context, slug domainCategory.Slug, loadTranslations bool) (*domainCategory.Category, error) {
	db := repository.GetDB(ctx, r.db)

	if loadTranslations {
		db = db.Preload("Translations")
	}

	var orm CategoryORM
	err := db.First(&orm, "slug = ?", slug.String()).Error
	if err != nil {
		translatedErr := repository.TranslateGormError(err)
		if errors.Is(translatedErr, gorm.ErrRecordNotFound) {
			return nil, domainCategory.ErrCategoryNotFound
		}
		return nil, fmt.Errorf("failed to query category by slug: %w", translatedErr)
	}

	return toDomain(&orm)
}

func (r *pgRepository) Update(ctx context.Context, c *domainCategory.Category) error {
	orm := toPersistence(c)
	db := repository.GetDB(ctx, r.db)

	oldVersion := orm.Version
	orm.Version++

	// The DB trigger will automatically update category_closure rows
	// when parent_id changes (rep parenting).
	result := db.Model(&CategoryORM{}).
		Where("id = ? AND version = ?", orm.ID, oldVersion).
		Updates(map[string]interface{}{
			"parent_id":  orm.ParentID,
			"slug":       orm.Slug,
			"status":     orm.Status,
			"cover_url":  orm.CoverURL,
			"position":   orm.Position,
			"updated_at": orm.UpdatedAt,
			"version":    orm.Version,
		})

	if result.Error != nil {
		translatedErr := repository.TranslateGormError(result.Error)
		if errors.Is(translatedErr, gorm.ErrDuplicatedKey) {
			return fmt.Errorf("%w: slug must be unique", domainCategory.ErrInvalidSlug)
		}
		return fmt.Errorf("failed to update category: %w", translatedErr)
	}

	if result.RowsAffected == 0 {
		return domainCategory.ErrConcurrentModification
	}

	return nil
}

func (r *pgRepository) SoftDelete(ctx context.Context, id domainCategory.CategoryID) error {
	db := repository.GetDB(ctx, r.db)

	// GORM with gorm.DeletedAt translates Delete into UPDATE SET deleted_at = NOW()
	result := db.Delete(&CategoryORM{}, "id = ?", uuid.UUID(id))
	if result.Error != nil {
		return fmt.Errorf("failed to soft delete category: %w", repository.TranslateGormError(result.Error))
	}

	if result.RowsAffected == 0 {
		return domainCategory.ErrCategoryNotFound
	}

	return nil
}

func (r *pgRepository) ExistsByID(ctx context.Context, id domainCategory.CategoryID) (bool, error) {
	db := repository.GetDB(ctx, r.db)

	var count int64
	err := db.Model(&CategoryORM{}).Where("id = ?", uuid.UUID(id)).Count(&count).Error
	if err != nil {
		return false, fmt.Errorf("failed to check category existence: %w", repository.TranslateGormError(err))
	}

	return count > 0, nil
}

// CheckDepthAndCycle verifies that reparenting targetNodeID under newParentID
// will not exceed max depth=2 and will not create a circular reference.
// The DB trigger enforces this strictly, but this check provides clean
// domain errors before hitting the DB.
func (r *pgRepository) CheckDepthAndCycle(ctx context.Context, targetNodeID, newParentID domainCategory.CategoryID) (bool, error) {
	db := repository.GetDB(ctx, r.db)

	// 1. Cycle check: is newParentID a descendant of targetNodeID?
	// If so, making targetNodeID a child of newParentID would create a cycle.
	var cycleExists bool
	err := db.Raw(`
        SELECT EXISTS(
            SELECT 1 FROM category_closure
            WHERE ancestor_id = ? AND descendant_id = ?
        )
    `, uuid.UUID(targetNodeID), uuid.UUID(newParentID)).Scan(&cycleExists).Error
	if err != nil {
		return false, fmt.Errorf("failed to check for cycle: %w", repository.TranslateGormError(err))
	}
	if cycleExists {
		return false, nil
	}

	// 2. Depth check: would the new depth exceed 2?
	// New parent's depth from root + 1 (for target node) + max subtree depth of target
	type depthCheck struct {
		ParentDepth  int `gorm:"column:parent_depth"`
		SubtreeDepth int `gorm:"column:subtree_depth"`
	}

	var result depthCheck
	err = db.Raw(`
        SELECT
            COALESCE((SELECT MAX(depth) FROM category_closure WHERE descendant_id = ?), -1) AS parent_depth,
            COALESCE((SELECT MAX(depth) FROM category_closure WHERE ancestor_id = ?), 0) AS subtree_depth
    `, uuid.UUID(newParentID), uuid.UUID(targetNodeID)).Scan(&result).Error
	if err != nil {
		return false, fmt.Errorf("failed to check depth: %w", repository.TranslateGormError(err))
	}

	// Total depth = parent_depth + 1 (for target) + subtree_depth
	// Max allowed depth = 2
	if result.ParentDepth+1+result.SubtreeDepth > 2 {
		return false, nil
	}

	return true, nil
}

func (r *pgRepository) UpsertTranslation(ctx context.Context, categoryID domainCategory.CategoryID, locale shared.Locale, title, description string) (*domainCategory.Category, error) {
	db := repository.GetDB(ctx, r.db)

	var desc *string
	if description != "" {
		desc = &description
	}

	transORM := CategoryTranslationORM{
		CategoryID:  uuid.UUID(categoryID),
		Locale:      string(locale),
		Title:       title,
		Description: desc,
	}

	err := db.Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "category_id"}, {Name: "locale"}},
		DoUpdates: clause.AssignmentColumns([]string{"title", "description"}),
	}).Create(&transORM).Error

	if err != nil {
		translatedErr := repository.TranslateGormError(err)
		if errors.Is(translatedErr, repository.ErrForeignKeyViolated) {
			return nil, domainCategory.ErrCategoryNotFound
		}
		return nil, fmt.Errorf("failed to upsert category translation: %w", translatedErr)
	}

	return r.GetByID(ctx, categoryID, true)
}

// ListPublishedTree uses a raw SQL query leveraging category_closure to build
// the hierarchy efficiently in one query, applying the lateral join for translations.
func (r *pgRepository) ListPublishedTree(ctx context.Context, preferredLocale shared.Locale) ([]*domainCategory.CategoryNode, error) {
	db := repository.GetDB(ctx, r.db)

	query := `
        SELECT
            c.id, c.parent_id, c.slug, c.status, c.cover_url, c.position,
            c.created_at, c.updated_at, c.version,
            t.locale AS trans_locale, t.title AS trans_title, t.description AS trans_description
        FROM categories c
        LEFT JOIN LATERAL (
            SELECT locale, title, description
            FROM category_translations
            WHERE category_id = c.id AND (locale = ? OR locale = 'en')
            ORDER BY CASE WHEN locale = ? THEN 1 ELSE 2 END
            LIMIT 1
        ) t ON true
        WHERE c.status = ? AND c.deleted_at IS NULL
        ORDER BY c.parent_id NULLS FIRST, c.position ASC
    `

	var flats []CategoryFlatORM
	err := db.Raw(query, string(preferredLocale), string(preferredLocale), string(domainCategory.CategoryStatusPublished)).Scan(&flats).Error
	if err != nil {
		return nil, fmt.Errorf("failed to query published category tree: %w", repository.TranslateGormError(err))
	}

	// Build tree structure using parent_id relationships
	nodes := make(map[uuid.UUID]*domainCategory.CategoryNode)
	var roots []*domainCategory.CategoryNode

	// First pass: create all nodes
	for i := range flats {
		flat := &flats[i]
		orm := flatToCategoryORM(flat)
		cat, err := toDomain(orm)
		if err != nil {
			return nil, err
		}
		nodes[flat.ID] = &domainCategory.CategoryNode{
			Category: cat,
			Children: []*domainCategory.CategoryNode{},
		}
	}

	// Second pass: build hierarchy
	for _, flat := range flats {
		node := nodes[flat.ID]
		if flat.ParentID == nil {
			roots = append(roots, node)
		} else {
			if parent, ok := nodes[*flat.ParentID]; ok {
				parent.Children = append(parent.Children, node)
			} else {
				// Parent not in the published tree; treat as root
				roots = append(roots, node)
			}
		}
	}

	return roots, nil
}
