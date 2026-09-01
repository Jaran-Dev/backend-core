package course

import (
	"context"
	"fmt"

	domainCourse "github.com/Jaran-Dev/backend-core/internal/domain/course"
)

type StudentCourseService struct {
	courseRepo domainCourse.Repository
}

func NewStudentCourseService(cr domainCourse.Repository) *StudentCourseService {
	return &StudentCourseService{
		courseRepo: cr,
	}
}

func (s *StudentCourseService) GetPublishedCourse(ctx context.Context, slug domainCourse.Slug, preferredLocale domainCourse.Locale) (*domainCourse.Course, domainCourse.Locale, error) {
	if !preferredLocale.IsValid() {
		preferredLocale = domainCourse.LocaleEnglish
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

func (s *StudentCourseService) ListPublishedCourses(ctx context.Context, params domainCourse.QueryParams, preferredLocale domainCourse.Locale) ([]*domainCourse.Course, int64, error) {
	if !preferredLocale.IsValid() {
		preferredLocale = domainCourse.LocaleEnglish
	}

	params.Status = string(domainCourse.CourseStatusPublished)

	courses, total, err := s.courseRepo.ListPublishedWithLocale(ctx, params, preferredLocale)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to list published courses: %w", err)
	}

	return courses, total, nil
}
