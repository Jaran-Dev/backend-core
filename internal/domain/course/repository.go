package course

import (
	"context"

	"github.com/Jaran-Dev/backend-core/internal/domain/shared"
	"github.com/google/uuid"
)

type QueryParams struct {
	Status     string    // "draft", "published", "archived" (empty string means all)
	Page       int       // 1-based indexing
	PageSize   int       // Number of items per page
	CategoryID uuid.UUID // Optional: Filter by category ID
}

type Repository interface {
	Create(ctx context.Context, c *Course) error
	GetByID(ctx context.Context, id CourseID, loadTranslations bool) (*Course, error)
	GetBySlug(ctx context.Context, slug Slug, loadTranslations bool) (*Course, error)
	Update(ctx context.Context, c *Course) error
	Delete(ctx context.Context, id CourseID) error
	List(ctx context.Context, params QueryParams) ([]*Course, int64, error)
	ExistsByID(ctx context.Context, id CourseID) (bool, error)
	UpsertTranslation(ctx context.Context, courseID CourseID, locale shared.Locale, title, description string) (*Course, error)
	GetPublishedBySlugWithLocale(ctx context.Context, slug Slug, preferredLocale shared.Locale) (*Course, shared.Locale, error)
	ListPublishedWithLocale(ctx context.Context, params QueryParams, preferredLocale shared.Locale) ([]*Course, int64, error)
	CountCoursesByCategoryID(ctx context.Context, categoryID uuid.UUID) (int64, error)
	GetPrimaryCategoryID(ctx context.Context, courseID CourseID) (uuid.UUID, error)
	AssignCategory(ctx context.Context, courseID CourseID, categoryID uuid.UUID, isPrimary bool, position int) error
	UnassignCategory(ctx context.Context, courseID CourseID, categoryID uuid.UUID) error
}
