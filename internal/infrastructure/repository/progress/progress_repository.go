package progress

import (
	"context"
	"errors"
	"fmt"
	"time"

	domainCourse "github.com/Jaran-Dev/backend-core/internal/domain/course"
	domainProgress "github.com/Jaran-Dev/backend-core/internal/domain/progress"
	domainUser "github.com/Jaran-Dev/backend-core/internal/domain/user"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

// --- Enrollment Repository ---

type pgEnrollmentRepository struct {
	pool *pgxpool.Pool
}

func NewEnrollmentRepository(pool *pgxpool.Pool) domainProgress.EnrollmentRepository {
	return &pgEnrollmentRepository{pool: pool}
}

func (r *pgEnrollmentRepository) Create(ctx context.Context, e *domainProgress.Enrollment) error {
	query := `
        INSERT INTO enrollments (id, user_id, course_id, enrolled_at)
        VALUES ($1, $2, $3, $4)
    `
	_, err := r.pool.Exec(ctx, query,
		e.ID(),
		uuid.UUID(e.UserID()),
		uuid.UUID(e.CourseID()),
		e.EnrolledAt(),
	)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return domainProgress.ErrAlreadyEnrolled
		}
		return fmt.Errorf("failed to create enrollment: %w", err)
	}
	return nil
}

func (r *pgEnrollmentRepository) GetByUserAndCourse(ctx context.Context, userID domainUser.UserID, courseID domainCourse.CourseID) (*domainProgress.Enrollment, error) {
	query := `SELECT id, user_id, course_id, enrolled_at FROM enrollments WHERE user_id = $1 AND course_id = $2`
	row := r.pool.QueryRow(ctx, query, uuid.UUID(userID), uuid.UUID(courseID))

	var (
		id         uuid.UUID
		uid        uuid.UUID
		cid        uuid.UUID
		enrolledAt time.Time
	)

	err := row.Scan(&id, &uid, &cid, &enrolledAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domainProgress.ErrEnrollmentNotFound
		}
		return nil, fmt.Errorf("failed to get enrollment: %w", err)
	}

	return domainProgress.UnmarshalEnrollment(id, domainUser.UserID(uid), domainCourse.CourseID(cid), enrolledAt)
}

func (r *pgEnrollmentRepository) ListByCourseID(ctx context.Context, courseID domainCourse.CourseID) ([]*domainProgress.Enrollment, error) {
	query := `SELECT id, user_id, course_id, enrolled_at FROM enrollments WHERE course_id = $1`
	rows, err := r.pool.Query(ctx, query, uuid.UUID(courseID))
	if err != nil {
		return nil, fmt.Errorf("failed to list enrollments: %w", err)
	}
	defer rows.Close()

	var enrollments []*domainProgress.Enrollment
	for rows.Next() {
		var id, uid, cid uuid.UUID
		var enrolledAt time.Time
		if err := rows.Scan(&id, &uid, &cid, &enrolledAt); err != nil {
			return nil, err
		}
		e, err := domainProgress.UnmarshalEnrollment(id, domainUser.UserID(uid), domainCourse.CourseID(cid), enrolledAt)
		if err != nil {
			return nil, err
		}
		enrollments = append(enrollments, e)
	}
	return enrollments, nil
}

func (r *pgEnrollmentRepository) ListByUserID(ctx context.Context, userID domainUser.UserID) ([]*domainProgress.Enrollment, error) {
	query := `SELECT id, user_id, course_id, enrolled_at FROM enrollments WHERE user_id = $1`
	rows, err := r.pool.Query(ctx, query, uuid.UUID(userID))
	if err != nil {
		return nil, fmt.Errorf("failed to list enrollments: %w", err)
	}
	defer rows.Close()

	var enrollments []*domainProgress.Enrollment
	for rows.Next() {
		var id, uid, cid uuid.UUID
		var enrolledAt time.Time
		if err := rows.Scan(&id, &uid, &cid, &enrolledAt); err != nil {
			return nil, err
		}
		e, err := domainProgress.UnmarshalEnrollment(id, domainUser.UserID(uid), domainCourse.CourseID(cid), enrolledAt)
		if err != nil {
			return nil, err
		}
		enrollments = append(enrollments, e)
	}
	return enrollments, nil
}

// --- Lesson Progress Repository ---

type pgLessonProgressRepository struct {
	pool *pgxpool.Pool
}

func NewLessonProgressRepository(pool *pgxpool.Pool) domainProgress.LessonProgressRepository {
	return &pgLessonProgressRepository{pool: pool}
}

func (r *pgLessonProgressRepository) Upsert(ctx context.Context, lp *domainProgress.LessonProgress) error {
	query := `
        INSERT INTO lesson_progress (id, user_id, lesson_id, watched_sec, completed, updated_at)
        VALUES ($1, $2, $3, $4, $5, $6)
        ON CONFLICT (user_id, lesson_id) DO UPDATE
        SET watched_sec = EXCLUDED.watched_sec, 
            completed = EXCLUDED.completed, 
            updated_at = EXCLUDED.updated_at
    `
	_, err := r.pool.Exec(ctx, query,
		lp.ID(),
		uuid.UUID(lp.UserID()),
		uuid.UUID(lp.LessonID()),
		lp.WatchedSec(),
		lp.IsCompleted(),
		lp.UpdatedAt(),
	)
	if err != nil {
		return fmt.Errorf("failed to upsert lesson progress: %w", err)
	}
	return nil
}

func (r *pgLessonProgressRepository) Get(ctx context.Context, userID domainUser.UserID, lessonID domainCourse.LessonID) (*domainProgress.LessonProgress, error) {
	query := `SELECT id, user_id, lesson_id, watched_sec, completed, updated_at FROM lesson_progress WHERE user_id = $1 AND lesson_id = $2`
	row := r.pool.QueryRow(ctx, query, uuid.UUID(userID), uuid.UUID(lessonID))

	var (
		id         uuid.UUID
		uid        uuid.UUID
		lid        uuid.UUID
		watchedSec int
		completed  bool
		updatedAt  time.Time
	)

	err := row.Scan(&id, &uid, &lid, &watchedSec, &completed, &updatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domainProgress.ErrProgressNotFound
		}
		return nil, fmt.Errorf("failed to get lesson progress: %w", err)
	}

	return domainProgress.UnmarshalLessonProgress(id, domainUser.UserID(uid), domainCourse.LessonID(lid), watchedSec, completed, updatedAt)
}

func (r *pgLessonProgressRepository) ListByUserAndCourse(ctx context.Context, userID domainUser.UserID, courseID domainCourse.CourseID) ([]*domainProgress.LessonProgress, error) {
	// We join with lessons and modules to ensure we only get progress for lessons belonging to this course
	query := `
        SELECT lp.id, lp.user_id, lp.lesson_id, lp.watched_sec, lp.completed, lp.updated_at 
        FROM lesson_progress lp
        JOIN lessons l ON lp.lesson_id = l.id
        JOIN modules m ON l.module_id = m.id
        WHERE lp.user_id = $1 AND m.course_id = $2
    `
	rows, err := r.pool.Query(ctx, query, uuid.UUID(userID), uuid.UUID(courseID))
	if err != nil {
		return nil, fmt.Errorf("failed to list lesson progress: %w", err)
	}
	defer rows.Close()

	var progresses []*domainProgress.LessonProgress
	for rows.Next() {
		var (
			id, uid, lid uuid.UUID
			watchedSec   int
			completed    bool
			updatedAt    time.Time
		)
		if err := rows.Scan(&id, &uid, &lid, &watchedSec, &completed, &updatedAt); err != nil {
			return nil, err
		}
		lp, err := domainProgress.UnmarshalLessonProgress(id, domainUser.UserID(uid), domainCourse.LessonID(lid), watchedSec, completed, updatedAt)
		if err != nil {
			return nil, err
		}
		progresses = append(progresses, lp)
	}
	return progresses, nil
}
