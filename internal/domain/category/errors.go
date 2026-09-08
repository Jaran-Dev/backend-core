package category

import "errors"

var (
	ErrCategoryNotFound          = errors.New("category not found")
	ErrInvalidSlug               = errors.New("invalid slug format: must be lowercase alphanumeric separated by hyphens")
	ErrInvalidLocale             = errors.New("invalid locale: must be en, ckb, or kmr")
	ErrInvalidPosition           = errors.New("invalid position: must be 0 or greater")
	ErrInvalidStateTransition    = errors.New("invalid state transition for the current category status")
	ErrMissingEnglishTranslation = errors.New("category must have an English translation to be published")
	ErrTranslationNotFound       = errors.New("translation not found for the given locale")
	ErrConcurrentModification    = errors.New("concurrent modification detected: the data was updated by another transaction")
	ErrCategoryHasCourses        = errors.New("category cannot be deleted or archived while courses are assigned to it")
	ErrMaxDepthExceeded          = errors.New("category hierarchy max depth exceeded")
	ErrCircularReference         = errors.New("circular reference detected in category hierarchy")
)
