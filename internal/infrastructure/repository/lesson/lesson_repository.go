package lesson

import (
	"context"
	"errors"
	"fmt"
	"time"

	domainCourse "github.com/Jaran-Dev/backend-core/internal/domain/course"
	domainLesson "github.com/Jaran-Dev/backend-core/internal/domain/course/lesson"
	"github.com/Jaran-Dev/backend-core/internal/infrastructure/repository"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

type pgRepository struct {
	pool *pgxpool.Pool
}

func NewRepository(pool *pgxpool.Pool) domainLesson.Repository {
	return &pgRepository{pool: pool}
}

func (r *pgRepository) Create(ctx context.Context, l *domainLesson.Lesson) error {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	asset := l.Asset()
	lessonQuery := `
        INSERT INTO lessons (id, module_id, position, access_type, asset_id, playback_id, duration_sec, status, created_at, updated_at)
        VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
    `
	_, err = tx.Exec(ctx, lessonQuery,
		uuid.UUID(l.ID()),
		uuid.UUID(l.ModuleID()),
		l.Position(),
		string(l.AccessType()),
		repository.NullableString(asset.AssetID),
		repository.NullableString(asset.PlaybackID),
		asset.DurationSec,
		string(l.Status()),
		l.CreatedAt(),
		l.UpdatedAt(),
	)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return fmt.Errorf("lesson position already exists in this module: %w", err)
		}
		return fmt.Errorf("failed to create lesson: %w", err)
	}

	transQuery := `INSERT INTO lesson_translations (lesson_id, locale, title, description) VALUES ($1, $2, $3, $4)`
	for _, t := range l.Translations() {
		_, err = tx.Exec(ctx, transQuery, uuid.UUID(t.LessonID()), string(t.Locale()), t.Title(), repository.NullableString(t.Description()))
		if err != nil {
			return fmt.Errorf("failed to create lesson translation: %w", err)
		}
	}

	subQuery := `INSERT INTO subtitles (id, lesson_id, locale, url) VALUES ($1, $2, $3, $4)`
	for _, s := range l.Subtitles() {
		_, err = tx.Exec(ctx, subQuery, uuid.New(), uuid.UUID(s.LessonID()), string(s.Locale()), s.URL())
		if err != nil {
			return fmt.Errorf("failed to create subtitle: %w", err)
		}
	}

	return tx.Commit(ctx)
}

func (r *pgRepository) GetByID(ctx context.Context, id domainCourse.LessonID) (*domainLesson.Lesson, error) {
	lessonQuery := `
        SELECT l.id, l.module_id, l.position, l.access_type, l.asset_id, l.playback_id, l.duration_sec, l.status, l.created_at, l.updated_at,
               t.locale, t.title, t.description
        FROM lessons l
        LEFT JOIN lesson_translations t ON l.id = t.lesson_id
        WHERE l.id = $1
    `
	rows, err := r.pool.Query(ctx, lessonQuery, uuid.UUID(id))
	if err != nil {
		return nil, fmt.Errorf("failed to query lesson: %w", err)
	}
	defer rows.Close()

	var translations []domainLesson.LessonTranslation
	var (
		lessonID    uuid.UUID
		moduleID    uuid.UUID
		position    int
		accessType  string
		assetID     *string
		playbackID  *string
		durationSec int
		status      string
		createdAt   time.Time
		updatedAt   time.Time
		found       = false
	)

	for rows.Next() {
		var (
			localeStr   *string
			title       *string
			description *string
		)
		err = rows.Scan(
			&lessonID, &moduleID, &position, &accessType, &assetID, &playbackID, &durationSec, &status, &createdAt, &updatedAt,
			&localeStr, &title, &description,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan lesson row: %w", err)
		}
		found = true

		if localeStr != nil && title != nil {
			t, err := domainLesson.UnmarshalLessonTranslation(domainCourse.LessonID(lessonID), domainCourse.Locale(*localeStr), *title, repository.DerefString(description))
			if err != nil {
				return nil, err
			}
			translations = append(translations, t)
		}
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating lesson rows: %w", err)
	}

	if !found {
		return nil, domainLesson.ErrLessonNotFound
	}

	subQuery := `SELECT locale, url FROM subtitles WHERE lesson_id = $1`
	subRows, err := r.pool.Query(ctx, subQuery, uuid.UUID(id))
	if err != nil {
		return nil, fmt.Errorf("failed to query subtitles: %w", err)
	}
	defer subRows.Close()

	var subtitles []domainLesson.Subtitle
	for subRows.Next() {
		var localeStr, urlStr string
		err = subRows.Scan(&localeStr, &urlStr)
		if err != nil {
			return nil, fmt.Errorf("failed to scan subtitle row: %w", err)
		}
		s, err := domainLesson.UnmarshalSubtitle(domainCourse.LessonID(lessonID), domainCourse.Locale(localeStr), urlStr)
		if err != nil {
			return nil, err
		}
		subtitles = append(subtitles, s)
	}

	if err = subRows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating subtitle rows: %w", err)
	}

	assetIDStr := repository.DerefString(assetID)
	playbackIDStr := repository.DerefString(playbackID)
	asset, err := domainLesson.UnmarshalVideoAsset(assetIDStr, playbackIDStr, durationSec, status)
	if err != nil {
		return nil, err
	}

	return domainLesson.UnmarshalLesson(
		lessonID, moduleID, position, accessType, asset, status, createdAt, updatedAt, translations, subtitles,
	)
}

func (r *pgRepository) GetByModuleID(ctx context.Context, moduleID domainCourse.ModuleID) ([]*domainLesson.Lesson, error) {
	query := `SELECT id FROM lessons WHERE module_id = $1 ORDER BY position ASC`
	rows, err := r.pool.Query(ctx, query, uuid.UUID(moduleID))
	if err != nil {
		return nil, fmt.Errorf("failed to query lessons by module: %w", err)
	}
	defer rows.Close()

	var lessons []*domainLesson.Lesson
	for rows.Next() {
		var id uuid.UUID
		if err := rows.Scan(&id); err != nil {
			return nil, err
		}
		l, err := r.GetByID(ctx, domainCourse.LessonID(id))
		if err != nil {
			return nil, err
		}
		lessons = append(lessons, l)
	}
	return lessons, nil
}

func (r *pgRepository) Update(ctx context.Context, l *domainLesson.Lesson) error {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	asset := l.Asset()
	lessonQuery := `
        UPDATE lessons 
        SET position = $1, access_type = $2, asset_id = $3, playback_id = $4, duration_sec = $5, status = $6, updated_at = $7
        WHERE id = $8
    `
	_, err = tx.Exec(ctx, lessonQuery,
		l.Position(),
		string(l.AccessType()),
		repository.NullableString(asset.AssetID),
		repository.NullableString(asset.PlaybackID),
		asset.DurationSec,
		string(l.Status()),
		l.UpdatedAt(),
		uuid.UUID(l.ID()),
	)
	if err != nil {
		return fmt.Errorf("failed to update lesson: %w", err)
	}

	// Atomic update for translations: delete and re-insert
	_, err = tx.Exec(ctx, `DELETE FROM lesson_translations WHERE lesson_id = $1`, uuid.UUID(l.ID()))
	if err != nil {
		return fmt.Errorf("failed to clear lesson translations: %w", err)
	}
	transQuery := `INSERT INTO lesson_translations (lesson_id, locale, title, description) VALUES ($1, $2, $3, $4)`
	for _, t := range l.Translations() {
		_, err = tx.Exec(ctx, transQuery, uuid.UUID(t.LessonID()), string(t.Locale()), t.Title(), repository.NullableString(t.Description()))
		if err != nil {
			return fmt.Errorf("failed to update lesson translation: %w", err)
		}
	}

	// Atomic update for subtitles: delete and re-insert
	_, err = tx.Exec(ctx, `DELETE FROM subtitles WHERE lesson_id = $1`, uuid.UUID(l.ID()))
	if err != nil {
		return fmt.Errorf("failed to clear subtitles: %w", err)
	}
	subQuery := `INSERT INTO subtitles (id, lesson_id, locale, url) VALUES ($1, $2, $3, $4)`
	for _, s := range l.Subtitles() {
		_, err = tx.Exec(ctx, subQuery, uuid.New(), uuid.UUID(s.LessonID()), string(s.Locale()), s.URL())
		if err != nil {
			return fmt.Errorf("failed to update subtitle: %w", err)
		}
	}

	return tx.Commit(ctx)
}

func (r *pgRepository) Delete(ctx context.Context, id domainCourse.LessonID) error {
	_, err := r.pool.Exec(ctx, `DELETE FROM lessons WHERE id = $1`, uuid.UUID(id))
	if err != nil {
		return fmt.Errorf("failed to delete lesson: %w", err)
	}
	return nil
}
