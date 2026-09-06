package lesson

import "github.com/Jaran-Dev/backend-core/internal/domain/shared"

type Subtitle struct {
	lessonID LessonID
	locale   shared.Locale
	url      string
}

func NewSubtitle(lessonID LessonID, locale shared.Locale, url string) Subtitle {
	return Subtitle{
		lessonID: lessonID,
		locale:   locale,
		url:      url,
	}
}

func (s Subtitle) LessonID() LessonID {
	return s.lessonID
}

func (s Subtitle) Locale() shared.Locale {
	return s.locale
}

func (s Subtitle) URL() string {
	return s.url
}
