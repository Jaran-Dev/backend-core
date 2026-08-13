package course

type CourseTranslation struct {
	courseID    CourseID
	locale      Locale
	title       string
	description string
}

func NewCourseTranslation(courseID CourseID, locale Locale, title, description string) CourseTranslation {
	return CourseTranslation{
		courseID:    courseID,
		locale:      locale,
		title:       title,
		description: description,
	}
}

func (t *CourseTranslation) CourseID() CourseID {
	return t.courseID
}

func (t *CourseTranslation) Locale() Locale {
	return t.locale
}

func (t *CourseTranslation) Title() string {
	return t.title
}

func (t *CourseTranslation) Description() string {
	return t.description
}
