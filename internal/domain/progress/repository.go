package progress

import (
	"context"

	"github.com/Jaran-Dev/backend-core/internal/domain/account"
	"github.com/Jaran-Dev/backend-core/internal/domain/course"
	"github.com/Jaran-Dev/backend-core/internal/domain/course/lesson"
)

type EnrollmentRepository interface {
	Create(ctx context.Context, e *Enrollment) error
	GetByUserAndCourse(ctx context.Context, accountID account.AccountID, courseID course.CourseID) (*Enrollment, error)
	ListByCourseID(ctx context.Context, courseID course.CourseID) ([]*Enrollment, error)
	ListByUserID(ctx context.Context, accountID account.AccountID) ([]*Enrollment, error)
}

type LessonProgressRepository interface {
	Upsert(ctx context.Context, lp *LessonProgress) error
	Get(ctx context.Context, accountID account.AccountID, lessonID lesson.LessonID) (*LessonProgress, error)
	ListByUserAndCourse(ctx context.Context, accountID account.AccountID, courseID course.CourseID) ([]*LessonProgress, error)
}
