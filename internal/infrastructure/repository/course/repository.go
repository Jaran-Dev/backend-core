package course

import (
	"context"
	"errors"
	"fmt"

	domainCourse "github.com/Jaran-Dev/backend-core/internal/domain/course"
	"github.com/Jaran-Dev/backend-core/internal/domain/shared"
	"github.com/Jaran-Dev/backend-core/internal/infrastructure/repository"

	"github.com/google/uuid"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type pgRepository struct {
	db *gorm.DB
}

func NewCourseRepository(db *gorm.DB) domainCourse.Repository {
	return &pgRepository{db: db}
}

func (r *pgRepository) Create(ctx context.Context, c *domainCourse.Course) error {
	orm := toPersistence(c)
	db := repository.GetDB(ctx, r.db)

	if err := db.Create(orm).Error; err != nil {
		translatedErr := repository.TranslateGormError(err)
		if errors.Is(translatedErr, gorm.ErrDuplicatedKey) {
			return fmt.Errorf("%w: slug must be unique", domainCourse.ErrInvalidSlug)
		}
		return fmt.Errorf("failed to create course in database: %w", translatedErr)
	}
	return nil
}

func (r *pgRepository) GetByID(ctx context.Context, id domainCourse.CourseID, loadTranslations bool) (*domainCourse.Course, error) {
	db := repository.GetDB(ctx, r.db)

	if loadTranslations {
		db = db.Preload("Translations")
	}

	var orm CourseORM
	err := db.First(&orm, "id = ?", uuid.UUID(id)).Error
	if err != nil {
		translatedErr := repository.TranslateGormError(err)
		if errors.Is(translatedErr, gorm.ErrRecordNotFound) {
			return nil, domainCourse.ErrCourseNotFound
		}
		return nil, fmt.Errorf("failed to query course by id: %w", translatedErr)
	}

	return toDomain(&orm)
}

func (r *pgRepository) GetBySlug(ctx context.Context, slug domainCourse.Slug, loadTranslations bool) (*domainCourse.Course, error) {
	db := repository.GetDB(ctx, r.db)

	if loadTranslations {
		db = db.Preload("Translations")
	}

	var orm CourseORM
	err := db.First(&orm, "slug = ?", slug.String()).Error
	if err != nil {
		translatedErr := repository.TranslateGormError(err)
		if errors.Is(translatedErr, gorm.ErrRecordNotFound) {
			return nil, domainCourse.ErrCourseNotFound
		}
		return nil, fmt.Errorf("failed to query course by slug: %w", translatedErr)
	}

	return toDomain(&orm)
}

func (r *pgRepository) Update(ctx context.Context, c *domainCourse.Course) error {
	orm := toPersistence(c)
	db := repository.GetDB(ctx, r.db)

	oldVersion := orm.Version
	orm.Version++

	result := db.Model(&CourseORM{}).
		Where("id = ? AND version = ?", orm.ID, oldVersion).
		Updates(map[string]interface{}{
			"slug":       orm.Slug,
			"status":     orm.Status,
			"cover_url":  orm.CoverURL,
			"updated_at": orm.UpdatedAt,
			"version":    orm.Version,
		})

	if result.Error != nil {
		translatedErr := repository.TranslateGormError(result.Error)
		if errors.Is(translatedErr, gorm.ErrDuplicatedKey) {
			return fmt.Errorf("%w: slug must be unique", domainCourse.ErrInvalidSlug)
		}
		return fmt.Errorf("failed to update course: %w", translatedErr)
	}

	if result.RowsAffected == 0 {
		return domainCourse.ErrConcurrentModification
	}

	return nil
}

func (r *pgRepository) ExistsByID(ctx context.Context, id domainCourse.CourseID) (bool, error) {
	db := repository.GetDB(ctx, r.db)

	var count int64
	err := db.Model(&CourseORM{}).Where("id = ?", uuid.UUID(id)).Count(&count).Error
	if err != nil {
		return false, fmt.Errorf("failed to check course existence: %w", repository.TranslateGormError(err))
	}

	return count > 0, nil
}

func (r *pgRepository) Delete(ctx context.Context, id domainCourse.CourseID) error {
	db := repository.GetDB(ctx, r.db)

	result := db.Delete(&CourseORM{}, "id = ?", uuid.UUID(id))
	if result.Error != nil {
		return fmt.Errorf("failed to delete course: %w", repository.TranslateGormError(result.Error))
	}

	if result.RowsAffected == 0 {
		return domainCourse.ErrCourseNotFound
	}

	return nil
}

func (r *pgRepository) UpsertTranslation(ctx context.Context, courseID domainCourse.CourseID, locale shared.Locale, title, description string) (*domainCourse.Course, error) {
	db := repository.GetDB(ctx, r.db)

	transORM := CourseTranslationORM{
		CourseID:    uuid.UUID(courseID),
		Locale:      string(locale),
		Title:       title,
		Description: repository.StringPtr(description),
	}

	err := db.Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "course_id"}, {Name: "locale"}},
		DoUpdates: clause.AssignmentColumns([]string{"title", "description"}),
	}).Create(&transORM).Error

	if err != nil {
		translatedErr := repository.TranslateGormError(err)
		if errors.Is(translatedErr, repository.ErrForeignKeyViolated) {
			return nil, domainCourse.ErrCourseNotFound
		}
		return nil, fmt.Errorf("failed to upsert translation: %w", translatedErr)
	}

	return r.GetByID(ctx, courseID, true)
}

func (r *pgRepository) GetPublishedBySlugWithLocale(ctx context.Context, slug domainCourse.Slug, preferredLocale shared.Locale) (*domainCourse.Course, shared.Locale, error) {
	db := repository.GetDB(ctx, r.db)

	query := `
        SELECT 
            c.id, c.slug, c.status, c.cover_url, c.created_by, c.created_at, c.updated_at, c.version,
            t.locale AS trans_locale, t.title AS trans_title, t.description AS trans_description
        FROM courses c
        LEFT JOIN LATERAL (
            SELECT locale, title, description
            FROM course_translations
            WHERE course_id = c.id AND (locale = ? OR locale = 'en')
            ORDER BY CASE WHEN locale = ? THEN 1 ELSE 2 END
            LIMIT 1
        ) t ON true
        WHERE c.slug = ? AND c.status = ?
    `

	var flat CourseFlatORM
	err := db.Raw(query, string(preferredLocale), string(preferredLocale), slug.String(), string(domainCourse.CourseStatusPublished)).Scan(&flat).Error
	if err != nil {
		translatedErr := repository.TranslateGormError(err)
		if errors.Is(translatedErr, gorm.ErrRecordNotFound) {
			return nil, "", domainCourse.ErrCourseNotFound
		}
		return nil, "", fmt.Errorf("failed to query published course by slug: %w", translatedErr)
	}

	orm := flatToCourseORM(&flat)
	course, err := toDomain(orm)
	if err != nil {
		return nil, "", err
	}

	resolvedLocale := shared.LocaleEnglish
	if flat.TransLocale != nil {
		resolvedLocale = shared.Locale(*flat.TransLocale)
	}

	return course, resolvedLocale, nil
}

func (r *pgRepository) ListPublishedWithLocale(ctx context.Context, params domainCourse.QueryParams, preferredLocale shared.Locale) ([]*domainCourse.Course, int64, error) {
	db := repository.GetDB(ctx, r.db)

	baseQuery := db.Model(&CourseORM{}).Where("status = ?", string(domainCourse.CourseStatusPublished))

	if params.CategoryID != uuid.Nil {
		baseQuery = baseQuery.Where("id IN (SELECT course_id FROM course_categories WHERE category_id = ?)", params.CategoryID)
	}

	var total int64
	if err := baseQuery.Count(&total).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to count published courses: %w", repository.TranslateGormError(err))
	}

	paginatedQuery := repository.ApplyPagination(baseQuery, params.Page, params.PageSize)

	query := db.Table("(?) AS c", paginatedQuery).
		Select(`c.id, c.slug, c.status, c.cover_url, c.created_by, c.created_at, c.updated_at, c.version,
                t.locale AS trans_locale, t.title AS trans_title, t.description AS trans_description`).
		Joins(`LEFT JOIN LATERAL (
            SELECT locale, title, description
            FROM course_translations
            WHERE course_id = c.id AND (locale = ? OR locale = 'en')
            ORDER BY CASE WHEN locale = ? THEN 1 ELSE 2 END
            LIMIT 1
        ) t ON true`, string(preferredLocale), string(preferredLocale))

	var flats []CourseFlatORM
	err := query.Scan(&flats).Error
	if err != nil {
		return nil, 0, fmt.Errorf("failed to list published courses: %w", repository.TranslateGormError(err))
	}

	courses := make([]*domainCourse.Course, len(flats))
	for i, flat := range flats {
		orm := flatToCourseORM(&flat)
		course, err := toDomain(orm)
		if err != nil {
			return nil, 0, err
		}
		courses[i] = course
	}

	return courses, total, nil
}

func (r *pgRepository) List(ctx context.Context, params domainCourse.QueryParams) ([]*domainCourse.Course, int64, error) {
	db := repository.GetDB(ctx, r.db)

	query := db.Model(&CourseORM{})

	if params.Status != "" {
		query = query.Where("status = ?", params.Status)
	}

	if params.CategoryID != uuid.Nil {
		query = query.Where("id IN (SELECT course_id FROM course_categories WHERE category_id = ?)", params.CategoryID)
	}

	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to count courses: %w", repository.TranslateGormError(err))
	}

	var orms []CourseORM
	paginatedQuery := repository.ApplyPagination(query, params.Page, params.PageSize)

	err := paginatedQuery.
		Preload("Translations").
		Order("created_at DESC").
		Find(&orms).Error

	if err != nil {
		return nil, 0, fmt.Errorf("failed to list courses: %w", repository.TranslateGormError(err))
	}

	courses := make([]*domainCourse.Course, len(orms))
	for i, orm := range orms {
		course, err := toDomain(&orm)
		if err != nil {
			return nil, 0, err
		}
		courses[i] = course
	}

	return courses, total, nil
}

func (r *pgRepository) CountCoursesByCategoryID(ctx context.Context, categoryID uuid.UUID) (int64, error) {
	db := repository.GetDB(ctx, r.db)

	var count int64
	err := db.Model(&CourseCategoryORM{}).Where("category_id = ?", categoryID).Count(&count).Error
	if err != nil {
		return 0, fmt.Errorf("failed to count courses by category id: %w", repository.TranslateGormError(err))
	}
	return count, nil
}

func (r *pgRepository) GetPrimaryCategoryID(ctx context.Context, courseID domainCourse.CourseID) (uuid.UUID, error) {
	db := repository.GetDB(ctx, r.db)

	var orm CourseCategoryORM
	err := db.Where("course_id = ? AND is_primary = ?", uuid.UUID(courseID), true).First(&orm).Error
	if err != nil {
		translatedErr := repository.TranslateGormError(err)
		if errors.Is(translatedErr, gorm.ErrRecordNotFound) {
			return uuid.Nil, nil
		}
		return uuid.Nil, fmt.Errorf("failed to get primary category id: %w", translatedErr)
	}
	return orm.CategoryID, nil
}

func (r *pgRepository) AssignCategory(ctx context.Context, courseID domainCourse.CourseID, categoryID uuid.UUID, isPrimary bool, position int) error {
	db := repository.GetDB(ctx, r.db)

	orm := CourseCategoryORM{
		CourseID:   uuid.UUID(courseID),
		CategoryID: categoryID,
		IsPrimary:  isPrimary,
		Position:   position,
	}

	err := db.Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "course_id"}, {Name: "category_id"}},
		DoUpdates: clause.AssignmentColumns([]string{"is_primary", "position"}),
	}).Create(&orm).Error

	if err != nil {
		translatedErr := repository.TranslateGormError(err)
		if errors.Is(translatedErr, gorm.ErrDuplicatedKey) {
			return domainCourse.ErrPrimaryCategoryAlreadyExists
		}
		if errors.Is(translatedErr, repository.ErrForeignKeyViolated) {
			return domainCourse.ErrCourseNotFound
		}
		return fmt.Errorf("failed to assign category to course: %w", translatedErr)
	}

	return nil
}

func (r *pgRepository) UnassignCategory(ctx context.Context, courseID domainCourse.CourseID, categoryID uuid.UUID) error {
	db := repository.GetDB(ctx, r.db)

	result := db.Where("course_id = ? AND category_id = ?", uuid.UUID(courseID), categoryID).
		Delete(&CourseCategoryORM{})

	if result.Error != nil {
		return fmt.Errorf("failed to unassign category from course: %w", repository.TranslateGormError(result.Error))
	}

	if result.RowsAffected == 0 {
		return domainCourse.ErrCourseNotFound
	}

	return nil
}
