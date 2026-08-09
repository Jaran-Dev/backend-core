package course

import "errors"

var (
	ErrCourseNotFound            = errors.New("course not found")
	ErrModuleNotFound            = errors.New("module not found")
	ErrInvalidSlug               = errors.New("invalid slug format: must be lowercase alphanumeric separated by hyphens")
	ErrInvalidLocale             = errors.New("invalid locale: must be en, ckb, or kmr")
	ErrMissingEnglishTranslation = errors.New("course must have an English translation to be published")
	ErrMissingModules            = errors.New("course must have at least one module to be published")
)
