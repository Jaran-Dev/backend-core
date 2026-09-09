package category

import (
	"regexp"

	"github.com/google/uuid"
)

type CategoryID uuid.UUID

func (id CategoryID) String() string { return uuid.UUID(id).String() }

type CategoryStatus string

const (
	CategoryStatusDraft     CategoryStatus = "draft"
	CategoryStatusPublished CategoryStatus = "published"
	CategoryStatusArchived  CategoryStatus = "archived"
)

func (s CategoryStatus) IsValid() bool {
	switch s {
	case CategoryStatusDraft, CategoryStatusPublished, CategoryStatusArchived:
		return true
	}
	return false
}

type Position int

func NewPosition(p int) (Position, error) {
	if p < 0 {
		return 0, ErrInvalidPosition
	}
	return Position(p), nil
}

type Slug string

var slugRegex = regexp.MustCompile(`^[a-z0-9]+(?:-[a-z0-9]+)*$`)

func NewSlug(s string) (Slug, error) {
	if !slugRegex.MatchString(s) {
		return "", ErrInvalidSlug
	}
	return Slug(s), nil
}

func (s Slug) String() string { return string(s) }
