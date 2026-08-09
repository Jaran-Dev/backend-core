package lesson

import "github.com/Jaran-Dev/backend-core/internal/domain/course"

type Subtitle struct {
	lessonID course.LessonID
	locale   course.Locale
	url      string
}

func NewSubtitle(lessonID course.LessonID, locale course.Locale, url string) Subtitle {
	return Subtitle{
		lessonID: lessonID,
		locale:   locale,
		url:      url,
	}
}

func (s Subtitle) LessonID() course.LessonID {
	return s.lessonID
}

func (s Subtitle) Locale() course.Locale {
	return s.locale
}

func (s Subtitle) URL() string {
	return s.url
}
