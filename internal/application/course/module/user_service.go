package module

import (
	"context"
	"fmt"

	domainCourse "github.com/Jaran-Dev/backend-core/internal/domain/course"
	domainModule "github.com/Jaran-Dev/backend-core/internal/domain/course/module"
)

type StudentModuleService struct {
	moduleRepo domainModule.Repository
	courseRepo domainCourse.Repository
}

func NewStudentModuleService(mr domainModule.Repository, cr domainCourse.Repository) *StudentModuleService {
	return &StudentModuleService{
		moduleRepo: mr,
		courseRepo: cr,
	}
}

func (s *StudentModuleService) ListModulesByCourse(ctx context.Context, courseID domainCourse.CourseID) ([]*domainModule.Module, error) {
	course, err := s.courseRepo.GetByID(ctx, courseID, false)
	if err != nil {
		return nil, err
	}

	if course.Status() != domainCourse.CourseStatusPublished {
		return nil, domainCourse.ErrCourseNotFound
	}

	modules, err := s.moduleRepo.ListByCourseID(ctx, courseID)
	if err != nil {
		return nil, fmt.Errorf("failed to list modules by course: %w", err)
	}

	return modules, nil
}
