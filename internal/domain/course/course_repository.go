package course

import "context"

type Repository interface {
	Create(ctx context.Context, c *Course) error
	GetByID(ctx context.Context, id CourseID) (*Course, error)
	GetBySlug(ctx context.Context, slug Slug) (*Course, error)
	Update(ctx context.Context, c *Course) error
	Delete(ctx context.Context, id CourseID) error
}
