package lesson

import "github.com/Jaran-Dev/backend-core/internal/domain/course"

type LessonTranslation struct {
	lessonID    course.LessonID
	locale      course.Locale
	title       string
	description string
}

func NewLessonTranslation(lessonID course.LessonID, locale course.Locale, title, description string) LessonTranslation {
	return LessonTranslation{
		lessonID:    lessonID,
		locale:      locale,
		title:       title,
		description: description,
	}
}

func (t *LessonTranslation) LessonID() course.LessonID {
	return t.lessonID
}

func (t *LessonTranslation) Locale() course.Locale {
	return t.locale
}

func (t *LessonTranslation) Title() string {
	return t.title
}

func (t *LessonTranslation) Description() string {
	return t.description
}
