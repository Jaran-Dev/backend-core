package course

import (
	"regexp"

	"github.com/google/uuid"
)

type CourseID uuid.UUID

type ModuleID uuid.UUID

type LessonID uuid.UUID

func (id CourseID) String() string {
	return uuid.UUID(id).String()
}

func (id ModuleID) String() string {
	return uuid.UUID(id).String()
}

func (id LessonID) String() string {
	return uuid.UUID(id).String()
}

type Locale string

const (
	LocaleEnglish         Locale = "en"
	LocaleSoraniKurdish   Locale = "ckb"
	LocaleKurmanjiKurdish Locale = "kmr"
)

func (l Locale) IsValid() bool {
	switch l {
	case LocaleEnglish, LocaleSoraniKurdish, LocaleKurmanjiKurdish:
		return true
	}
	return false
}

type CourseStatus string

const (
	CourseStatusDraft     CourseStatus = "draft"
	CourseStatusPublished CourseStatus = "published"
	CourseStatusArchived  CourseStatus = "archived"
)

type LessonAccessType string

const (
	AccessFree    LessonAccessType = "free"    // Available to everyone, no login required
	AccessPreview LessonAccessType = "preview" // Available to logged-in users
	AccessPremium LessonAccessType = "premium" // Available only to subscribed users
)

func (a LessonAccessType) IsValid() bool {
	switch a {
	case AccessFree, AccessPreview, AccessPremium:
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
