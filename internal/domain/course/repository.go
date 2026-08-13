package course

import "context"

// QueryParams encapsulates pagination and filtering for listing courses.
type QueryParams struct {
	Status   string // "draft", "published", "archived" (empty string means all)
	Page     int    // 1-based indexing
	PageSize int    // Number of items per page
}

type Repository interface {
	Create(ctx context.Context, c *Course) error
	GetByID(ctx context.Context, id CourseID, loadTranslations bool) (*Course, error)
	GetBySlug(ctx context.Context, slug Slug, loadTranslations bool) (*Course, error)
	Update(ctx context.Context, c *Course) error
	Delete(ctx context.Context, id CourseID) error
	List(ctx context.Context, params QueryParams) ([]*Course, int64, error)
}
