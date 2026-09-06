package preferences

type Locale string

const (
	LocaleEnglish         Locale = "en"
	LocaleSoraniKurdish   Locale = "ckb"
	LocaleKurmanjiKurdish Locale = "kmr"
)

func (l Locale) IsValid() bool {
	switch l {
	case LocaleEnglish, LocaleSoraniKurdish, LocaleKurmanjiKurdish:
		return true
	}
	return false
}

func NewLocale(s string) (Locale, error) {
	l := Locale(s)
	if !l.IsValid() {
		return "", ErrInvalidLocale
	}
	return l, nil
}

type Theme string

const (
	ThemeLight  Theme = "light"
	ThemeDark   Theme = "dark"
	ThemeSystem Theme = "system"
)

func NewTheme(s string) (Theme, error) {
	t := Theme(s)
	switch t {
	case ThemeLight, ThemeDark, ThemeSystem:
		return t, nil
	}
	return "", ErrInvalidTheme
}

type PlaybackQuality string

const (
	QualityAuto  PlaybackQuality = "auto"
	Quality1080p PlaybackQuality = "1080p"
	Quality720p  PlaybackQuality = "720p"
	Quality480p  PlaybackQuality = "480p"
)

func NewPlaybackQuality(s string) (PlaybackQuality, error) {
	q := PlaybackQuality(s)
	switch q {
	case QualityAuto, Quality1080p, Quality720p, Quality480p:
		return q, nil
	}
	return "", ErrInvalidPlaybackQuality
}

type DigestFrequency string

const (
	DigestDaily   DigestFrequency = "daily"
	DigestWeekly  DigestFrequency = "weekly"
	DigestMonthly DigestFrequency = "monthly"
	DigestNever   DigestFrequency = "never"
)

func NewDigestFrequency(s string) (DigestFrequency, error) {
	f := DigestFrequency(s)
	switch f {
	case DigestDaily, DigestWeekly, DigestMonthly, DigestNever:
		return f, nil
	}
	return "", ErrInvalidDigestFrequency
}

type NotificationSettings struct {
	CourseProgress bool `json:"course_progress"`
	NewLessons     bool `json:"new_lessons"`
	LiveEvents     bool `json:"live_events"`
	Marketing      bool `json:"marketing"`
	SecurityAlerts bool `json:"security_alerts"`
	WeeklySummary  bool `json:"weekly_summary"`
	ProductUpdates bool `json:"product_updates"`
}
