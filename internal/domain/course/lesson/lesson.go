package lesson

import (
	"errors"
	"time"

	"github.com/Jaran-Dev/backend-core/internal/domain/course"
	"github.com/Jaran-Dev/backend-core/internal/domain/course/module"
	"github.com/Jaran-Dev/backend-core/internal/domain/shared"

	"github.com/google/uuid"
)

type Lesson struct {
	id         LessonID
	moduleID   module.ModuleID
	position   int
	accessType LessonAccessType
	asset      VideoAsset
	status     LessonVideoStatus // Derived from asset status
	createdAt  time.Time
	updatedAt  time.Time

	translations []LessonTranslation
	subtitles    []Subtitle
}

func NewLesson(moduleID module.ModuleID, position int) *Lesson {
	now := time.Now()
	return &Lesson{
		id:         LessonID(uuid.New()),
		moduleID:   moduleID,
		position:   position,
		accessType: AccessPreview, // Default to preview according to MVP baseline
		status:     VideoStatusUploading,
		createdAt:  now,
		updatedAt:  now,
	}
}

func (l *Lesson) ID() LessonID {
	return l.id
}

func (l *Lesson) ModuleID() module.ModuleID {
	return l.moduleID
}

func (l *Lesson) Position() int {
	return l.position
}

func (l *Lesson) AccessType() LessonAccessType {
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

func (l *Lesson) UpdateAccessType(access LessonAccessType) error {
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

func (l *Lesson) AddOrUpdateTranslation(locale shared.Locale, title, desc string) error {
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

func (l *Lesson) AddSubtitle(locale shared.Locale, url string) error {
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
	if !LessonAccessType(accessTypeStr).IsValid() {
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
		id:           LessonID(id),
		moduleID:     module.ModuleID(moduleID),
		position:     position,
		accessType:   LessonAccessType(accessTypeStr),
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

func UnmarshalLessonTranslation(lessonID LessonID, locale shared.Locale, title, description string) (LessonTranslation, error) {
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

func UnmarshalSubtitle(lessonID LessonID, locale shared.Locale, url string) (Subtitle, error) {
	if !locale.IsValid() {
		return Subtitle{}, course.ErrInvalidLocale
	}
	return Subtitle{
		lessonID: lessonID,
		locale:   locale,
		url:      url,
	}, nil
}
