package course

import (
	"context"
	"errors"
	"fmt"

	domainCourse "github.com/Jaran-Dev/backend-core/internal/domain/course"
	"github.com/Jaran-Dev/backend-core/internal/infrastructure/repository"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

type pgModuleRepository struct {
	db repository.DBTX
}

func NewModuleRepository(db repository.DBTX) domainCourse.ModuleRepository {
	return &pgModuleRepository{db: db}
}

func (r *pgModuleRepository) Create(ctx context.Context, m *domainCourse.Module) error {
	query := `
        INSERT INTO modules (id, course_id, position, title)
        VALUES ($1, $2, $3, $4)
    `
	_, err := r.db.Exec(ctx, query,
		uuid.UUID(m.ID()),
		uuid.UUID(m.CourseID()),
		m.Position(),
		m.Title(),
	)
	if err != nil {
		var pgErr *pgconn.PgError
		// 23505 is unique_violation (e.g., duplicate position or ID)
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return fmt.Errorf("module position already exists in this course: %w", err)
		}
		return fmt.Errorf("failed to create module: %w", err)
	}
	return nil
}

func (r *pgModuleRepository) GetByID(ctx context.Context, id domainCourse.ModuleID) (*domainCourse.Module, error) {
	query := `
        SELECT id, course_id, position, title
        FROM modules
        WHERE id = $1
    `
	row := r.db.QueryRow(ctx, query, uuid.UUID(id))
	return r.scanModule(row)
}

func (r *pgModuleRepository) ListByCourseID(ctx context.Context, courseID domainCourse.CourseID) ([]*domainCourse.Module, error) {
	query := `
        SELECT id, course_id, position, title
        FROM modules
        WHERE course_id = $1
        ORDER BY position ASC
    `
	rows, err := r.db.Query(ctx, query, uuid.UUID(courseID))
	if err != nil {
		return nil, fmt.Errorf("failed to query modules: %w", err)
	}
	defer rows.Close()

	var modules []*domainCourse.Module
	for rows.Next() {
		m, err := r.scanModule(rows)
		if err != nil {
			return nil, err
		}
		modules = append(modules, m)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating module rows: %w", err)
	}

	return modules, nil
}

func (r *pgModuleRepository) Update(ctx context.Context, m *domainCourse.Module) error {
	query := `
        UPDATE modules
        SET position = $1, title = $2
        WHERE id = $3
    `
	_, err := r.db.Exec(ctx, query,
		m.Position(),
		m.Title(),
		uuid.UUID(m.ID()),
	)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return fmt.Errorf("module position already exists in this course: %w", err)
		}
		return fmt.Errorf("failed to update module: %w", err)
	}
	return nil
}

func (r *pgModuleRepository) Delete(ctx context.Context, id domainCourse.ModuleID) error {
	query := `DELETE FROM modules WHERE id = $1`
	_, err := r.db.Exec(ctx, query, uuid.UUID(id))
	if err != nil {
		return fmt.Errorf("failed to delete module: %w", err)
	}
	return nil
}

type rowScanner interface {
	Scan(dest ...interface{}) error
}

func (r *pgModuleRepository) scanModule(row rowScanner) (*domainCourse.Module, error) {
	var (
		id       uuid.UUID
		courseID uuid.UUID
		position int
		title    string
	)

	err := row.Scan(
		&id,
		&courseID,
		&position,
		&title,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domainCourse.ErrModuleNotFound
		}
		return nil, fmt.Errorf("failed to scan module row: %w", err)
	}

	return domainCourse.UnmarshalModule(id, domainCourse.CourseID(courseID), position, title)
}
