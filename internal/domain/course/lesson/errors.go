package lesson

import "errors"

var (
	ErrLessonNotFound      = errors.New("lesson not found")
	ErrInvalidAccessType   = errors.New("invalid access type: must be free, preview, or premium")
	ErrLessonVideoNotReady = errors.New("lesson video is not ready for playback")
)
