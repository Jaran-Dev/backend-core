package lesson

import "github.com/google/uuid"

type LessonID uuid.UUID

func (id LessonID) String() string {
	return uuid.UUID(id).String()
}

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
