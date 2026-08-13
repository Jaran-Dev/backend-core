package course

import "errors"

var (
	ErrCourseNotFound            = errors.New("course not found")
	ErrInvalidSlug               = errors.New("invalid slug format: must be lowercase alphanumeric separated by hyphens")
	ErrInvalidLocale             = errors.New("invalid locale: must be en, ckb, or kmr")
	ErrMissingEnglishTranslation = errors.New("course must have an English translation to be published")
)
