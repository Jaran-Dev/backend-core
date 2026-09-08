package category

import (
	"context"

	"github.com/Jaran-Dev/backend-core/internal/domain/shared"
)

type Repository interface {
	Create(ctx context.Context, c *Category) error
	GetByID(ctx context.Context, id CategoryID, loadTranslations bool) (*Category, error)
	GetBySlug(ctx context.Context, slug Slug, loadTranslations bool) (*Category, error)
	Update(ctx context.Context, c *Category) error
	SoftDelete(ctx context.Context, id CategoryID) error

	ExistsByID(ctx context.Context, id CategoryID) (bool, error)

	// CheckDepthAndCycle is used to verify reparenting logic before applying it.
	// Although DB triggers enforce this strictly, checking in the domain layer
	// provides clean domain errors before hitting the DB.
	CheckDepthAndCycle(ctx context.Context, targetNodeID, newParentID CategoryID) (bool, error)

	UpsertTranslation(ctx context.Context, categoryID CategoryID, locale shared.Locale, title, description string) (*Category, error)

	// ListPublishedTree returns all published categories in a flat list;
	// the application layer or infra layer can build the tree structure using parent_id.
	ListPublishedTree(ctx context.Context, preferredLocale shared.Locale) ([]*Category, error)
}

type CategoryNode struct {
	Category *Category
	Children []*CategoryNode
}
