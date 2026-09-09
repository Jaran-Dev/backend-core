package category

import (
	"context"
	"fmt"

	domainCategory "github.com/Jaran-Dev/backend-core/internal/domain/category"
	domainCourse "github.com/Jaran-Dev/backend-core/internal/domain/course"
	"github.com/Jaran-Dev/backend-core/internal/domain/shared"

	"github.com/google/uuid"
)

type TransactionManager interface {
	Do(ctx context.Context, fn func(ctx context.Context) error) error
}

type AdminCategoryService struct {
	categoryRepo domainCategory.Repository
	courseRepo   domainCourse.Repository
	txManager    TransactionManager
}

func NewAdminCategoryService(
	cr domainCategory.Repository,
	courseRepo domainCourse.Repository,
	txm TransactionManager,
) *AdminCategoryService {
	return &AdminCategoryService{
		categoryRepo: cr,
		courseRepo:   courseRepo,
		txManager:    txm,
	}
}

func (s *AdminCategoryService) Create(ctx context.Context, slugStr, coverURL string, parentID *uuid.UUID) (*domainCategory.Category, error) {
	slug, err := domainCategory.NewSlug(slugStr)
	if err != nil {
		return nil, err
	}

	var pID *domainCategory.CategoryID
	if parentID != nil {
		// Verify parent exists
		pid := domainCategory.CategoryID(*parentID)
		exists, err := s.categoryRepo.ExistsByID(ctx, pid)
		if err != nil {
			return nil, fmt.Errorf("failed to verify parent category: %w", err)
		}
		if !exists {
			return nil, domainCategory.ErrCategoryNotFound
		}
		pID = &pid
	}

	category := domainCategory.NewCategory(slug, coverURL, pID)

	if err := s.categoryRepo.Create(ctx, category); err != nil {
		return nil, err
	}

	return category, nil
}

func (s *AdminCategoryService) UpdateMetadata(ctx context.Context, id domainCategory.CategoryID, slugStr *string, coverURL *string) (*domainCategory.Category, error) {
	category, err := s.categoryRepo.GetByID(ctx, id, false)
	if err != nil {
		return nil, err
	}

	if slugStr != nil || coverURL != nil {
		slug := category.Slug()
		if slugStr != nil {
			slug, err = domainCategory.NewSlug(*slugStr)
			if err != nil {
				return nil, err
			}
		}

		cover := category.CoverURL()
		if coverURL != nil {
			cover = *coverURL
		}

		category.UpdateMetadata(slug, cover)
	}

	if err := s.categoryRepo.Update(ctx, category); err != nil {
		return nil, err
	}

	return category, nil
}

func (s *AdminCategoryService) UpsertTranslation(ctx context.Context, id domainCategory.CategoryID, locale shared.Locale, title, description string) (*domainCategory.Category, error) {
	if !locale.IsValid() {
		return nil, domainCategory.ErrInvalidLocale
	}

	exists, err := s.categoryRepo.ExistsByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to check category existence: %w", err)
	}
	if !exists {
		return nil, domainCategory.ErrCategoryNotFound
	}

	return s.categoryRepo.UpsertTranslation(ctx, id, locale, title, description)
}

func (s *AdminCategoryService) Publish(ctx context.Context, id domainCategory.CategoryID) (*domainCategory.Category, error) {
	category, err := s.categoryRepo.GetByID(ctx, id, true)
	if err != nil {
		return nil, err
	}

	if err := category.Publish(); err != nil {
		return nil, err
	}

	if err := s.categoryRepo.Update(ctx, category); err != nil {
		return nil, err
	}

	return category, nil
}

func (s *AdminCategoryService) Archive(ctx context.Context, id domainCategory.CategoryID) (*domainCategory.Category, error) {
	category, err := s.categoryRepo.GetByID(ctx, id, false)
	if err != nil {
		return nil, err
	}

	if err := category.Archive(); err != nil {
		return nil, err
	}

	if err := s.categoryRepo.Update(ctx, category); err != nil {
		return nil, err
	}

	return category, nil
}

func (s *AdminCategoryService) SoftDelete(ctx context.Context, id domainCategory.CategoryID) error {
	count, err := s.courseRepo.CountCoursesByCategoryID(ctx, uuid.UUID(id))
	if err != nil {
		return fmt.Errorf("failed to check courses for category: %w", err)
	}
	if count > 0 {
		return domainCategory.ErrCategoryHasCourses
	}

	category, err := s.categoryRepo.GetByID(ctx, id, false)
	if err != nil {
		return err
	}

	if err := category.SoftDelete(); err != nil {
		return err
	}

	return s.categoryRepo.SoftDelete(ctx, id)
}

func (s *AdminCategoryService) Reparent(ctx context.Context, id domainCategory.CategoryID, newParentID *uuid.UUID) (*domainCategory.Category, error) {
	// If newParentID is nil, we're detaching (making root)
	if newParentID != nil {
		parentID := domainCategory.CategoryID(*newParentID)

		// Verify parent exists
		parentExists, err := s.categoryRepo.ExistsByID(ctx, parentID)
		if err != nil {
			return nil, fmt.Errorf("failed to verify parent category: %w", err)
		}
		if !parentExists {
			return nil, domainCategory.ErrCategoryNotFound
		}

		// Check depth and cycle constraints
		valid, err := s.categoryRepo.CheckDepthAndCycle(ctx, id, parentID)
		if err != nil {
			return nil, err
		}
		if !valid {
			return nil, domainCategory.ErrInvalidDepth
		}

		var pID *domainCategory.CategoryID = &parentID
		return s.reparentCategory(ctx, id, pID)
	}

	// Detaching from parent (becoming root)
	return s.reparentCategory(ctx, id, nil)
}

func (s *AdminCategoryService) reparentCategory(ctx context.Context, id domainCategory.CategoryID, newParentID *domainCategory.CategoryID) (*domainCategory.Category, error) {
	var updatedCategory *domainCategory.Category

	err := s.txManager.Do(ctx, func(txCtx context.Context) error {
		category, err := s.categoryRepo.GetByID(txCtx, id, false)
		if err != nil {
			return err
		}

		category.Reparent(newParentID)

		if err := s.categoryRepo.Update(txCtx, category); err != nil {
			return err
		}

		updatedCategory = category
		return nil
	})

	if err != nil {
		return nil, err
	}
	return updatedCategory, nil
}
