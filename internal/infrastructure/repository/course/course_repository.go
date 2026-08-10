package course

import (
	"context"
	"errors"
	"fmt"
	"time"

	domainAdmin "github.com/Jaran-Dev/backend-core/internal/domain/admin"
	domainCourse "github.com/Jaran-Dev/backend-core/internal/domain/course"
	"github.com/Jaran-Dev/backend-core/internal/infrastructure/repository"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

type pgRepository struct {
	pool *pgxpool.Pool
}

func NewRepository(pool *pgxpool.Pool) domainCourse.Repository {
	return &pgRepository{pool: pool}
}

func (r *pgRepository) Create(ctx context.Context, c *domainCourse.Course) error {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	courseQuery := `
        INSERT INTO courses (id, slug, status, cover_url, created_by, created_at, updated_at)
        VALUES ($1, $2, $3, $4, $5, $6, $7)
    `
	_, err = tx.Exec(ctx, courseQuery,
		uuid.UUID(c.ID()),
		c.Slug().String(),
		string(c.Status()),
		repository.NullableString(c.CoverURL()),
		uuid.UUID(c.CreatedBy()),
		c.CreatedAt(),
		c.UpdatedAt(),
	)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return fmt.Errorf("course with this slug already exists: %w", err)
		}
		return fmt.Errorf("failed to create course: %w", err)
	}

	translationQuery := `
        INSERT INTO course_translations (course_id, locale, title, description)
        VALUES ($1, $2, $3, $4)
    `
	for _, t := range c.Translations() {
		_, err := tx.Exec(ctx, translationQuery,
			uuid.UUID(t.CourseID()),
			string(t.Locale()),
			t.Title(),
			repository.NullableString(t.Description()),
		)
		if err != nil {
			return fmt.Errorf("failed to create course translation: %w", err)
		}
	}

	return tx.Commit(ctx)
}

func (r *pgRepository) GetByID(ctx context.Context, id domainCourse.CourseID) (*domainCourse.Course, error) {
	query := `
        SELECT 
            c.id, c.slug, c.status, c.cover_url, c.created_by, c.created_at, c.updated_at,
            t.locale, t.title, t.description
        FROM courses c
        LEFT JOIN course_translations t ON c.id = t.course_id
        WHERE c.id = $1
    `
	rows, err := r.pool.Query(ctx, query, uuid.UUID(id))
	if err != nil {
		return nil, fmt.Errorf("failed to query course: %w", err)
	}
	defer rows.Close()

	return r.scanCourseRows(rows)
}

func (r *pgRepository) GetBySlug(ctx context.Context, slug domainCourse.Slug) (*domainCourse.Course, error) {
	query := `
        SELECT 
            c.id, c.slug, c.status, c.cover_url, c.created_by, c.created_at, c.updated_at,
            t.locale, t.title, t.description
        FROM courses c
        LEFT JOIN course_translations t ON c.id = t.course_id
        WHERE c.slug = $1
    `
	rows, err := r.pool.Query(ctx, query, slug.String())
	if err != nil {
		return nil, fmt.Errorf("failed to query course by slug: %w", err)
	}
	defer rows.Close()

	return r.scanCourseRows(rows)
}

func (r *pgRepository) Update(ctx context.Context, c *domainCourse.Course) error {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	courseQuery := `
        UPDATE courses 
        SET slug = $1, status = $2, cover_url = $3, updated_at = $4
        WHERE id = $5
    `
	_, err = tx.Exec(ctx, courseQuery,
		c.Slug().String(),
		string(c.Status()),
		repository.NullableString(c.CoverURL()),
		c.UpdatedAt(),
		uuid.UUID(c.ID()),
	)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return fmt.Errorf("course with this slug already exists: %w", err)
		}
		return fmt.Errorf("failed to update course: %w", err)
	}

	// For translations, the simplest atomic approach in Postgres is to delete existing
	// and insert the current state of the aggregate. This ensures the DB perfectly matches memory.
	_, err = tx.Exec(ctx, `DELETE FROM course_translations WHERE course_id = $1`, uuid.UUID(c.ID()))
	if err != nil {
		return fmt.Errorf("failed to clear course translations: %w", err)
	}

	translationQuery := `
        INSERT INTO course_translations (course_id, locale, title, description)
        VALUES ($1, $2, $3, $4)
    `
	for _, t := range c.Translations() {
		_, err := tx.Exec(ctx, translationQuery,
			uuid.UUID(t.CourseID()),
			string(t.Locale()),
			t.Title(),
			repository.NullableString(t.Description()),
		)
		if err != nil {
			return fmt.Errorf("failed to update course translation: %w", err)
		}
	}

	return tx.Commit(ctx)
}

func (r *pgRepository) Delete(ctx context.Context, id domainCourse.CourseID) error {
	query := `DELETE FROM courses WHERE id = $1`
	_, err := r.pool.Exec(ctx, query, uuid.UUID(id))
	if err != nil {
		return fmt.Errorf("failed to delete course: %w", err)
	}
	return nil
}

func (r *pgRepository) scanCourseRows(rows pgx.Rows) (*domainCourse.Course, error) {
	var translations []domainCourse.CourseTranslation

	var (
		id        uuid.UUID
		slugStr   string
		statusStr string
		coverURL  *string
		createdBy uuid.UUID
		createdAt time.Time
		updatedAt time.Time
		found     = false
	)

	for rows.Next() {
		var (
			localeStr   *string // Nullable because of LEFT JOIN
			title       *string // Nullable
			description *string // Nullable
		)

		err := rows.Scan(
			&id, &slugStr, &statusStr, &coverURL, &createdBy, &createdAt, &updatedAt,
			&localeStr, &title, &description,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan course row: %w", err)
		}

		found = true

		// If the LEFT JOIN found a translation, reconstruct and append it
		if localeStr != nil && title != nil {
			locale := domainCourse.Locale(*localeStr)
			descStr := ""
			if description != nil {
				descStr = *description
			}

			t, err := domainCourse.UnmarshalCourseTranslation(domainCourse.CourseID(id), locale, *title, descStr)
			if err != nil {
				return nil, err
			}
			translations = append(translations, t)
		}
	}

	if !found {
		return nil, domainCourse.ErrCourseNotFound
	}

	coverStr := ""
	if coverURL != nil {
		coverStr = *coverURL
	}

	return domainCourse.UnmarshalCourse(
		id,
		slugStr,
		statusStr,
		coverStr,
		domainAdmin.AdminID(createdBy),
		createdAt,
		updatedAt,
		translations,
	)
}
