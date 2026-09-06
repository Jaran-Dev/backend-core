package preferences

import "errors"

var (
	ErrPreferencesNotFound    = errors.New("preferences not found")
	ErrConcurrentModification = errors.New("concurrent modification detected: preferences were updated by another transaction")

	// Validation Errors
	ErrInvalidLocale          = errors.New("invalid locale: must be en, ckb, or kmr")
	ErrInvalidTheme           = errors.New("invalid theme: must be light, dark, or system")
	ErrInvalidPlaybackQuality = errors.New("invalid playback quality: must be auto, 1080p, 720p, or 480p")
	ErrInvalidDigestFrequency = errors.New("invalid digest frequency: must be daily, weekly, monthly, or never")
)
