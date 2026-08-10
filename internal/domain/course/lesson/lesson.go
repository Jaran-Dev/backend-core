package lesson

import (
	"errors"
	"time"

	"github.com/Jaran-Dev/backend-core/internal/domain/course"

	"github.com/google/uuid"
)

type Lesson struct {
	id         course.LessonID
	moduleID   course.ModuleID
	position   int
	accessType course.LessonAccessType
	asset      VideoAsset
	status     LessonVideoStatus // Derived from asset status
	createdAt  time.Time
	updatedAt  time.Time

	translations []LessonTranslation
	subtitles    []Subtitle
}

func NewLesson(moduleID course.ModuleID, position int) *Lesson {
	now := time.Now()
	return &Lesson{
		id:         course.LessonID(uuid.New()),
		moduleID:   moduleID,
		position:   position,
		accessType: course.AccessPreview, // Default to preview according to MVP baseline
		status:     VideoStatusUploading,
		createdAt:  now,
		updatedAt:  now,
	}
}

func (l *Lesson) ID() course.LessonID {
	return l.id
}

func (l *Lesson) ModuleID() course.ModuleID {
	return l.moduleID
}

func (l *Lesson) Position() int {
	return l.position
}

func (l *Lesson) AccessType() course.LessonAccessType {
	return l.accessType
}

func (l *Lesson) Asset() VideoAsset {
	return l.asset
}

func (l *Lesson) Status() LessonVideoStatus {
	return l.status
}

func (l *Lesson) CreatedAt() time.Time {
	return l.createdAt
}

func (l *Lesson) UpdatedAt() time.Time {
	return l.updatedAt
}

func (l *Lesson) Translations() []LessonTranslation {
	return l.translations
}

func (l *Lesson) Subtitles() []Subtitle {
	return l.subtitles
}

func (l *Lesson) UpdateAccessType(access course.LessonAccessType) error {
	if !access.IsValid() {
		return ErrInvalidAccessType
	}

	l.accessType = access
	l.updatedAt = time.Now()
	return nil
}

func (l *Lesson) AssignVideoAsset(assetID string) {
	l.asset = VideoAsset{
		AssetID: assetID,
		Status:  VideoStatusUploading,
	}
	l.updatedAt = time.Now()
}

func (l *Lesson) MarkVideoReady(playbackID string, durationSec int) {
	l.asset.PlaybackID = playbackID
	l.asset.DurationSec = durationSec
	l.asset.Status = VideoStatusReady
	l.status = VideoStatusReady
	l.updatedAt = time.Now()
}

func (l *Lesson) MarkVideoError() {
	l.asset.Status = VideoStatusError
	l.status = VideoStatusError
	l.updatedAt = time.Now()
}

func (l *Lesson) AddOrUpdateTranslation(locale course.Locale, title, desc string) error {
	if !locale.IsValid() {
		return course.ErrInvalidLocale
	}

	for i, t := range l.translations {
		if t.locale == locale {
			l.translations[i].title = title
			l.translations[i].description = desc
			l.updatedAt = time.Now()
			return nil
		}
	}
	l.translations = append(l.translations, LessonTranslation{
		lessonID:    l.id,
		locale:      locale,
		title:       title,
		description: desc,
	})
	l.updatedAt = time.Now()
	return nil
}

func (l *Lesson) AddSubtitle(locale course.Locale, url string) error {
	if !locale.IsValid() {
		return course.ErrInvalidLocale
	}

	for i, s := range l.subtitles {
		if s.locale == locale {
			l.subtitles[i].url = url
			l.updatedAt = time.Now()
			return nil
		}
	}
	l.subtitles = append(l.subtitles, Subtitle{
		lessonID: l.id,
		locale:   locale,
		url:      url,
	})
	l.updatedAt = time.Now()
	return nil
}

func UnmarshalLesson(
	id uuid.UUID,
	moduleID uuid.UUID,
	position int,
	accessTypeStr string,
	asset VideoAsset,
	statusStr string,
	createdAt time.Time,
	updatedAt time.Time,
	translations []LessonTranslation,
	subtitles []Subtitle,
) (*Lesson, error) {
	if !course.LessonAccessType(accessTypeStr).IsValid() {
		return nil, ErrInvalidAccessType
	}

	var status LessonVideoStatus
	switch statusStr {
	case "uploading":
		status = VideoStatusUploading
	case "ready":
		status = VideoStatusReady
	case "error":
		status = VideoStatusError
	default:
		return nil, errors.New("invalid lesson status in database")
	}

	return &Lesson{
		id:           course.LessonID(id),
		moduleID:     course.ModuleID(moduleID),
		position:     position,
		accessType:   course.LessonAccessType(accessTypeStr),
		asset:        asset,
		status:       status,
		createdAt:    createdAt,
		updatedAt:    updatedAt,
		translations: translations,
		subtitles:    subtitles,
	}, nil
}

func UnmarshalVideoAsset(assetID, playbackID string, durationSec int, statusStr string) (VideoAsset, error) {
	var status LessonVideoStatus
	switch statusStr {
	case "uploading":
		status = VideoStatusUploading
	case "ready":
		status = VideoStatusReady
	case "error":
		status = VideoStatusError
	default:
		return VideoAsset{}, errors.New("invalid video status in database")
	}

	return VideoAsset{
		AssetID:     assetID,
		PlaybackID:  playbackID,
		DurationSec: durationSec,
		Status:      status,
	}, nil
}

func UnmarshalLessonTranslation(lessonID course.LessonID, locale course.Locale, title, description string) (LessonTranslation, error) {
	if !locale.IsValid() {
		return LessonTranslation{}, course.ErrInvalidLocale
	}
	return LessonTranslation{
		lessonID:    lessonID,
		locale:      locale,
		title:       title,
		description: description,
	}, nil
}

func UnmarshalSubtitle(lessonID course.LessonID, locale course.Locale, url string) (Subtitle, error) {
	if !locale.IsValid() {
		return Subtitle{}, course.ErrInvalidLocale
	}
	return Subtitle{
		lessonID: lessonID,
		locale:   locale,
		url:      url,
	}, nil
}
