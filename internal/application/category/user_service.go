package category

import (
	"context"
	"fmt"

	domainCategory "github.com/Jaran-Dev/backend-core/internal/domain/category"
	domainCourse "github.com/Jaran-Dev/backend-core/internal/domain/course"
	"github.com/Jaran-Dev/backend-core/internal/domain/shared"

	"github.com/google/uuid"
)

type StudentCategoryService struct {
	categoryRepo domainCategory.Repository
	courseRepo   domainCourse.Repository
}

func NewStudentCategoryService(
	cr domainCategory.Repository,
	courseRepo domainCourse.Repository,
) *StudentCategoryService {
	return &StudentCategoryService{
		categoryRepo: cr,
		courseRepo:   courseRepo,
	}
}

func (s *StudentCategoryService) GetTree(ctx context.Context, preferredLocale shared.Locale) ([]*domainCategory.CategoryNode, error) {
	if !preferredLocale.IsValid() {
		preferredLocale = shared.LocaleEnglish
	}

	return s.categoryRepo.ListPublishedTree(ctx, preferredLocale)
}

func (s *StudentCategoryService) ListCoursesByCategory(ctx context.Context, slugStr string, preferredLocale shared.Locale, page, pageSize int) ([]*domainCourse.Course, int64, error) {
	if !preferredLocale.IsValid() {
		preferredLocale = shared.LocaleEnglish
	}

	slug, err := domainCategory.NewSlug(slugStr)
	if err != nil {
		return nil, 0, err
	}

	category, err := s.categoryRepo.GetBySlug(ctx, slug, false)
	if err != nil {
		return nil, 0, err
	}

	if category.Status() != domainCategory.CategoryStatusPublished {
		return nil, 0, domainCategory.ErrCategoryNotFound
	}

	params := domainCourse.QueryParams{
		Page:       page,
		PageSize:   pageSize,
		Status:     string(domainCourse.CourseStatusPublished),
		CategoryID: uuid.UUID(category.ID()),
	}

	courses, total, err := s.courseRepo.ListPublishedWithLocale(ctx, params, preferredLocale)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to list courses by category: %w", err)
	}

	return courses, total, nil
}
