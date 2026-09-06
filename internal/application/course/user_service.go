package course

import (
	"context"
	"fmt"

	domainCourse "github.com/Jaran-Dev/backend-core/internal/domain/course"
	"github.com/Jaran-Dev/backend-core/internal/domain/shared"
)

type StudentCourseService struct {
	courseRepo domainCourse.Repository
}

func NewStudentCourseService(cr domainCourse.Repository) *StudentCourseService {
	return &StudentCourseService{
		courseRepo: cr,
	}
}

func (s *StudentCourseService) GetPublishedCourse(ctx context.Context, slug domainCourse.Slug, preferredLocale shared.Locale) (*domainCourse.Course, shared.Locale, error) {
	if !preferredLocale.IsValid() {
		preferredLocale = shared.LocaleEnglish
	}

	course, resolvedLocale, err := s.courseRepo.GetPublishedBySlugWithLocale(ctx, slug, preferredLocale)
	if err != nil {
		return nil, "", err
	}

	if course.Status() != domainCourse.CourseStatusPublished {
		return nil, "", domainCourse.ErrCourseNotFound
	}

	return course, resolvedLocale, nil
}

func (s *StudentCourseService) ListPublishedCourses(ctx context.Context, params domainCourse.QueryParams, preferredLocale shared.Locale) ([]*domainCourse.Course, int64, error) {
	if !preferredLocale.IsValid() {
		preferredLocale = shared.LocaleEnglish
	}

	params.Status = string(domainCourse.CourseStatusPublished)

	courses, total, err := s.courseRepo.ListPublishedWithLocale(ctx, params, preferredLocale)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to list published courses: %w", err)
	}

	return courses, total, nil
}
