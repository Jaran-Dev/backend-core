package progress

import (
	"errors"
	"time"

	"github.com/Jaran-Dev/backend-core/internal/domain/account"
	"github.com/Jaran-Dev/backend-core/internal/domain/course"
	"github.com/Jaran-Dev/backend-core/internal/domain/course/lesson"
	"github.com/google/uuid"
)

type Enrollment struct {
	id         uuid.UUID
	accountID  account.AccountID
	courseID   course.CourseID
	enrolledAt time.Time
}

func NewEnrollment(accountID account.AccountID, courseID course.CourseID) *Enrollment {
	return &Enrollment{
		id:         uuid.New(),
		accountID:  accountID,
		courseID:   courseID,
		enrolledAt: time.Now(),
	}
}

func (e *Enrollment) ID() uuid.UUID {
	return e.id
}

func (e *Enrollment) UserID() account.AccountID {
	return e.accountID
}

func (e *Enrollment) CourseID() course.CourseID {
	return e.courseID
}

func (e *Enrollment) EnrolledAt() time.Time {
	return e.enrolledAt
}

type LessonProgress struct {
	id         uuid.UUID
	accountID  account.AccountID
	lessonID   lesson.LessonID
	watchedSec int
	completed  bool
	updatedAt  time.Time
}

func NewLessonProgress(accountID account.AccountID, lessonID lesson.LessonID) *LessonProgress {
	return &LessonProgress{
		id:         uuid.New(),
		accountID:  accountID,
		lessonID:   lessonID,
		watchedSec: 0,
		completed:  false,
		updatedAt:  time.Now(),
	}
}

func (lp *LessonProgress) ID() uuid.UUID {
	return lp.id
}

func (lp *LessonProgress) UserID() account.AccountID {
	return lp.accountID
}

func (lp *LessonProgress) LessonID() lesson.LessonID {
	return lp.lessonID
}

func (lp *LessonProgress) WatchedSec() int {
	return lp.watchedSec
}

func (lp *LessonProgress) IsCompleted() bool {
	return lp.completed
}

func (lp *LessonProgress) UpdatedAt() time.Time {
	return lp.updatedAt
}

func (lp *LessonProgress) UpdateProgress(watchedSec int, totalDurationSec int) {
	if watchedSec > lp.watchedSec {
		lp.watchedSec = watchedSec
	}

	if totalDurationSec > 0 && watchedSec >= (totalDurationSec*9/10) {
		lp.completed = true
	}

	lp.updatedAt = time.Now()
}

func (lp *LessonProgress) MarkComplete() {
	lp.completed = true
	lp.updatedAt = time.Now()
}

func UnmarshalEnrollment(id uuid.UUID, accountID account.AccountID, courseID course.CourseID, enrolledAt time.Time) (*Enrollment, error) {
	if id == uuid.Nil {
		return nil, errors.New("cannot unmarshal enrollment: id is nil")
	}
	return &Enrollment{
		id:         id,
		accountID:  accountID,
		courseID:   courseID,
		enrolledAt: enrolledAt,
	}, nil
}

func UnmarshalLessonProgress(id uuid.UUID, accountID account.AccountID, lessonID lesson.LessonID, watchedSec int, completed bool, updatedAt time.Time) (*LessonProgress, error) {
	if id == uuid.Nil {
		return nil, errors.New("cannot unmarshal lesson progress: id is nil")
	}
	if watchedSec < 0 {
		return nil, errors.New("cannot unmarshal lesson progress: watchedSec cannot be negative")
	}

	return &LessonProgress{
		id:         id,
		accountID:  accountID,
		lessonID:   lessonID,
		watchedSec: watchedSec,
		completed:  completed,
		updatedAt:  updatedAt,
	}, nil
}
