package course

import (
	"regexp"

	"github.com/google/uuid"
)

type CourseID uuid.UUID

func (id CourseID) String() string {
	return uuid.UUID(id).String()
}

type CourseStatus string

const (
	CourseStatusDraft     CourseStatus = "draft"
	CourseStatusPublished CourseStatus = "published"
	CourseStatusArchived  CourseStatus = "archived"
)

func (s CourseStatus) IsValid() bool {
	switch s {
	case CourseStatusDraft, CourseStatusPublished, CourseStatusArchived:
		return true
	}
	return false
}

type Slug string

var slugRegex = regexp.MustCompile(`^[a-z0-9]+(?:-[a-z0-9]+)*$`)

func NewSlug(s string) (Slug, error) {
	if !slugRegex.MatchString(s) {
		return "", ErrInvalidSlug
	}
	return Slug(s), nil
}

func (s Slug) String() string {
	return string(s)
}
