package course

import (
	"context"
	"fmt"

	domainAdmin "github.com/Jaran-Dev/backend-core/internal/domain/admin"
	domainCourse "github.com/Jaran-Dev/backend-core/internal/domain/course"
	domainModule "github.com/Jaran-Dev/backend-core/internal/domain/course/module"
)

type AdminCourseService struct {
	courseRepo domainCourse.Repository
	moduleRepo domainModule.Repository
}

func NewAdminCourseService(cr domainCourse.Repository, mr domainModule.Repository) *AdminCourseService {
	return &AdminCourseService{
		courseRepo: cr,
		moduleRepo: mr,
	}
}

func (s *AdminCourseService) Create(ctx context.Context, slugStr, coverURL string, adminID domainAdmin.AdminID) (*domainCourse.Course, error) {
	slug, err := domainCourse.NewSlug(slugStr)
	if err != nil {
		return nil, err
	}

	course := domainCourse.NewCourse(slug, coverURL, adminID)

	if err := s.courseRepo.Create(ctx, course); err != nil {
		return nil, fmt.Errorf("failed to create course: %w", err)
	}

	return course, nil
}

func (s *AdminCourseService) UpdateMetadata(ctx context.Context, courseID domainCourse.CourseID, slugStr *string, coverURL *string) (*domainCourse.Course, error) {
	course, err := s.courseRepo.GetByID(ctx, courseID, false)
	if err != nil {
		return nil, err
	}

	if slugStr != nil {
		slug, err := domainCourse.NewSlug(*slugStr)
		if err != nil {
			return nil, err
		}
		course.UpdateMetadata(slug, course.CoverURL())
	}
	if coverURL != nil {
		course.UpdateMetadata(course.Slug(), *coverURL)
	}

	if err := s.courseRepo.Update(ctx, course); err != nil {
		return nil, err
	}

	return course, nil
}

func (s *AdminCourseService) UpsertTranslation(ctx context.Context, courseID domainCourse.CourseID, locale domainCourse.Locale, title, description string) (*domainCourse.Course, error) {
	if !locale.IsValid() {
		return nil, domainCourse.ErrInvalidLocale
	}

	exists, err := s.courseRepo.ExistsByID(ctx, courseID)
	if err != nil {
		return nil, fmt.Errorf("failed to check course existence: %w", err)
	}
	if !exists {
		return nil, domainCourse.ErrCourseNotFound
	}

	course, err := s.courseRepo.UpsertTranslation(ctx, courseID, locale, title, description)
	if err != nil {
		return nil, err
	}

	return course, nil
}

func (s *AdminCourseService) Publish(ctx context.Context, courseID domainCourse.CourseID) (*domainCourse.Course, error) {
	course, err := s.courseRepo.GetByID(ctx, courseID, true)
	if err != nil {
		return nil, err
	}

	count, err := s.moduleRepo.CountByCourseID(ctx, courseID)
	if err != nil {
		return nil, fmt.Errorf("failed to verify modules for publication: %w", err)
	}
	if count == 0 {
		return nil, domainCourse.ErrMissingModules
	}

	if err := course.Publish(); err != nil {
		return nil, err
	}

	if err := s.courseRepo.Update(ctx, course); err != nil {
		return nil, err
	}

	return course, nil
}

func (s *AdminCourseService) Archive(ctx context.Context, courseID domainCourse.CourseID) (*domainCourse.Course, error) {
	course, err := s.courseRepo.GetByID(ctx, courseID, false)
	if err != nil {
		return nil, err
	}

	if err := course.Archive(); err != nil {
		return nil, err
	}

	if err := s.courseRepo.Update(ctx, course); err != nil {
		return nil, err
	}

	return course, nil
}
