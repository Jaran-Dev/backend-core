package course

import "errors"

var (
	ErrCourseNotFound               = errors.New("course not found")
	ErrInvalidSlug                  = errors.New("invalid slug format: must be lowercase alphanumeric separated by hyphens")
	ErrInvalidLocale                = errors.New("invalid locale: must be en, ckb, or kmr")
	ErrMissingEnglishTranslation    = errors.New("course must have an English translation to be published")
	ErrTranslationNotFound          = errors.New("translation not found for the given locale")
	ErrInvalidStateTransition       = errors.New("invalid state transition for the current course status")
	ErrMissingModules               = errors.New("course must have at least one module to be published")
	ErrConcurrentModification       = errors.New("concurrent modification detected: the data was updated by another transaction")
	ErrPrimaryCategoryNotPublished  = errors.New("course primary category must be published")
	ErrPrimaryCategoryNotAssigned   = errors.New("course must have a primary category assigned to be published")
	ErrPrimaryCategoryAlreadyExists = errors.New("course already has a primary category assigned")
)
