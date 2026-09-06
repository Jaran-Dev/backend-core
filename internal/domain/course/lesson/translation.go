package lesson

import "github.com/Jaran-Dev/backend-core/internal/domain/shared"

type LessonTranslation struct {
	lessonID    LessonID
	locale      shared.Locale
	title       string
	description string
}

func NewLessonTranslation(lessonID LessonID, locale shared.Locale, title, description string) LessonTranslation {
	return LessonTranslation{
		lessonID:    lessonID,
		locale:      locale,
		title:       title,
		description: description,
	}
}

func (t *LessonTranslation) LessonID() LessonID {
	return t.lessonID
}

func (t *LessonTranslation) Locale() shared.Locale {
	return t.locale
}

func (t *LessonTranslation) Title() string {
	return t.title
}

func (t *LessonTranslation) Description() string {
	return t.description
}
