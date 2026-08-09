package progress

import (
	"context"

	"github.com/Jaran-Dev/backend-core/internal/domain/course"
	"github.com/Jaran-Dev/backend-core/internal/domain/user"
)

type EnrollmentRepository interface {
	Create(ctx context.Context, e *Enrollment) error
	GetByUserAndCourse(ctx context.Context, userID user.UserID, courseID course.CourseID) (*Enrollment, error)
	ListByCourseID(ctx context.Context, courseID course.CourseID) ([]*Enrollment, error)
	ListByUserID(ctx context.Context, userID user.UserID) ([]*Enrollment, error)
}

type LessonProgressRepository interface {
	Upsert(ctx context.Context, lp *LessonProgress) error
	Get(ctx context.Context, userID user.UserID, lessonID course.LessonID) (*LessonProgress, error)
	ListByUserAndCourse(ctx context.Context, userID user.UserID, courseID course.CourseID) ([]*LessonProgress, error)
}
