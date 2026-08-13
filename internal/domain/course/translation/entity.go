package translation

import "github.com/Jaran-Dev/backend-core/internal/domain/course"

type CourseTranslation struct {
	courseID    course.CourseID
	locale      course.Locale
	title       string
	description string
}

func NewCourseTranslation(courseID course.CourseID, locale course.Locale, title, description string) CourseTranslation {
	return CourseTranslation{
		courseID:    courseID,
		locale:      locale,
		title:       title,
		description: description,
	}
}

func (t *CourseTranslation) CourseID() course.CourseID {
	return t.courseID
}

func (t *CourseTranslation) Locale() course.Locale {
	return t.locale
}

func (t *CourseTranslation) Title() string {
	return t.title
}

func (t *CourseTranslation) Description() string {
	return t.description
}
