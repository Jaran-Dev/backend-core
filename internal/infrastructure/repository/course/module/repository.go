package module

import (
	"context"
	"fmt"
	"time"

	domainCourse "github.com/Jaran-Dev/backend-core/internal/domain/course"
	domainModule "github.com/Jaran-Dev/backend-core/internal/domain/course/module"
	"github.com/Jaran-Dev/backend-core/internal/infrastructure/repository"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type pgRepository struct {
	db *gorm.DB
}

func NewModuleRepository(db *gorm.DB) domainModule.Repository {
	return &pgRepository{db: db}
}

func (r *pgRepository) Create(ctx context.Context, m *domainModule.Module) error {
	orm := toPersistence(m)
	db := repository.GetDB(ctx, r.db)

	if err := db.Create(orm).Error; err != nil {
		translatedErr := repository.TranslateGormError(err)
		if translatedErr == gorm.ErrDuplicatedKey {
			return fmt.Errorf("%w: position %d is already taken in this course", domainModule.ErrInvalidPosition, orm.Position)
		}
		return fmt.Errorf("failed to create module in database: %w", translatedErr)
	}
	return nil
}

func (r *pgRepository) GetByID(ctx context.Context, id domainModule.ModuleID) (*domainModule.Module, error) {
	db := repository.GetDB(ctx, r.db)

	var orm ModuleORM
	err := db.First(&orm, "id = ?", uuid.UUID(id)).Error
	if err != nil {
		translatedErr := repository.TranslateGormError(err)
		if translatedErr == gorm.ErrRecordNotFound {
			return nil, domainModule.ErrModuleNotFound
		}
		return nil, fmt.Errorf("failed to query module by id: %w", translatedErr)
	}

	return toDomain(&orm)
}

func (r *pgRepository) ListByCourseID(ctx context.Context, courseID domainCourse.CourseID) ([]*domainModule.Module, error) {
	db := repository.GetDB(ctx, r.db)

	var orms []ModuleORM
	err := db.Where("course_id = ?", uuid.UUID(courseID)).Order("position ASC").Find(&orms).Error
	if err != nil {
		return nil, fmt.Errorf("failed to list modules by course id: %w", repository.TranslateGormError(err))
	}

	modules := make([]*domainModule.Module, len(orms))
	for i, orm := range orms {
		mod, err := toDomain(&orm)
		if err != nil {
			return nil, err
		}
		modules[i] = mod
	}
	return modules, nil
}

func (r *pgRepository) CountByCourseID(ctx context.Context, courseID domainCourse.CourseID) (int64, error) {
	db := repository.GetDB(ctx, r.db)

	var count int64
	err := db.Model(&ModuleORM{}).Where("course_id = ?", uuid.UUID(courseID)).Count(&count).Error
	if err != nil {
		return 0, fmt.Errorf("failed to count modules by course id: %w", repository.TranslateGormError(err))
	}
	return count, nil
}

func (r *pgRepository) Update(ctx context.Context, m *domainModule.Module) error {
	orm := toPersistence(m)
	db := repository.GetDB(ctx, r.db)

	oldVersion := orm.Version
	orm.Version++

	result := db.Model(&ModuleORM{}).
		Where("id = ? AND version = ?", orm.ID, oldVersion).
		Updates(map[string]interface{}{
			"title":      orm.Title,
			"position":   orm.Position,
			"updated_at": orm.UpdatedAt,
			"version":    orm.Version,
		})

	if result.Error != nil {
		translatedErr := repository.TranslateGormError(result.Error)
		if translatedErr == gorm.ErrDuplicatedKey {
			return fmt.Errorf("%w: position %d is already taken in this course", domainModule.ErrInvalidPosition, orm.Position)
		}
		return fmt.Errorf("failed to update module: %w", translatedErr)
	}

	if result.RowsAffected == 0 {
		return domainModule.ErrConcurrentModification
	}

	return nil
}

func (r *pgRepository) Delete(ctx context.Context, id domainModule.ModuleID) error {
	db := repository.GetDB(ctx, r.db)

	result := db.Delete(&ModuleORM{}, "id = ?", uuid.UUID(id))
	if result.Error != nil {
		return fmt.Errorf("failed to delete module: %w", repository.TranslateGormError(result.Error))
	}

	if result.RowsAffected == 0 {
		return domainModule.ErrModuleNotFound
	}

	return nil
}

// ==========================================
// 6. BULK REORDER (The Crown Jewel)
// ==========================================

// BulkReorder atomically updates the positions of multiple modules in a single query.
// Instead of looping `UPDATE` statements in Go (which is slow and causes lock contention),
// we use a PostgreSQL `CASE WHEN` clause.
//
// Example generated SQL:
// UPDATE modules SET position = CASE id
//
//	WHEN 'uuid-1' THEN 0
//	WHEN 'uuid-2' THEN 1
//
// END, updated_at = '...', version = version + 1
// WHERE course_id = 'course-uuid'
func (r *pgRepository) BulkReorder(ctx context.Context, courseID domainCourse.CourseID, positions map[domainModule.ModuleID]domainModule.Position) error {
	db := repository.GetDB(ctx, r.db)

	if len(positions) == 0 {
		return nil
	}

	caseSQL := "CASE id"
	var args []interface{}

	for modID, pos := range positions {
		caseSQL += " WHEN ? THEN ?"
		args = append(args, uuid.UUID(modID), int(pos))
	}
	caseSQL += " END"

	now := time.Now()
	args = append(args, now, uuid.UUID(courseID))

	query := fmt.Sprintf(`
        UPDATE modules 
        SET position = %s, 
            updated_at = ?, 
            version = version + 1 
        WHERE course_id = ?
    `, caseSQL)

	result := db.Exec(query, args...)
	if result.Error != nil {
		return fmt.Errorf("failed to bulk reorder modules: %w", repository.TranslateGormError(result.Error))
	}

	if result.RowsAffected != int64(len(positions)) {
		return fmt.Errorf("%w: expected to update %d modules, but updated %d", domainModule.ErrInvalidReorderPayload, len(positions), result.RowsAffected)
	}

	return nil
}
