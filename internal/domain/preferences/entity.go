package preferences

import (
	"time"

	"github.com/Jaran-Dev/backend-core/internal/domain/account"
	"github.com/Jaran-Dev/backend-core/internal/domain/shared"
)

type UpdateParams struct {
	UILocale             *string
	Theme                *string
	PlaybackQuality      *string
	AutoplayNext         *bool
	SubtitleLanguage     *string
	SubtitlesEnabled     *bool
	SkipIntro            *bool
	ReducedMotion        *bool
	EmailDigestFrequency *string
	NotificationSettings *NotificationSettings
}

type Preferences struct {
	accountID            account.AccountID
	uiLocale             shared.Locale
	theme                Theme
	playbackQuality      PlaybackQuality
	autoplayNext         bool
	subtitleLanguage     shared.Locale
	subtitlesEnabled     bool
	skipIntro            bool
	reducedMotion        bool
	emailDigestFrequency DigestFrequency
	notificationSettings NotificationSettings
	createdAt            time.Time
	updatedAt            time.Time
	version              int
}

func NewPreferences(accID account.AccountID) *Preferences {
	now := time.Now()
	return &Preferences{
		accountID:            accID,
		uiLocale:             shared.LocaleEnglish,
		theme:                ThemeSystem,
		playbackQuality:      QualityAuto,
		autoplayNext:         true,
		subtitleLanguage:     shared.LocaleEnglish,
		subtitlesEnabled:     true,
		skipIntro:            true,
		reducedMotion:        false,
		emailDigestFrequency: DigestWeekly,
		notificationSettings: NotificationSettings{
			CourseProgress: true,
			NewLessons:     true,
			LiveEvents:     true,
			Marketing:      false,
			SecurityAlerts: true,
			WeeklySummary:  true,
			ProductUpdates: true,
		},
		createdAt: now,
		updatedAt: now,
		version:   1,
	}
}

func (p *Preferences) AccountID() account.AccountID               { return p.accountID }
func (p *Preferences) UILocale() shared.Locale                    { return p.uiLocale }
func (p *Preferences) Theme() Theme                               { return p.theme }
func (p *Preferences) PlaybackQuality() PlaybackQuality           { return p.playbackQuality }
func (p *Preferences) AutoplayNext() bool                         { return p.autoplayNext }
func (p *Preferences) SubtitleLanguage() shared.Locale            { return p.subtitleLanguage }
func (p *Preferences) SubtitlesEnabled() bool                     { return p.subtitlesEnabled }
func (p *Preferences) SkipIntro() bool                            { return p.skipIntro }
func (p *Preferences) ReducedMotion() bool                        { return p.reducedMotion }
func (p *Preferences) EmailDigestFrequency() DigestFrequency      { return p.emailDigestFrequency }
func (p *Preferences) NotificationSettings() NotificationSettings { return p.notificationSettings }
func (p *Preferences) CreatedAt() time.Time                       { return p.createdAt }
func (p *Preferences) UpdatedAt() time.Time                       { return p.updatedAt }
func (p *Preferences) Version() int                               { return p.version }

func (p *Preferences) Update(params UpdateParams) error {
	if params.UILocale != nil {
		l, err := shared.NewLocale(*params.UILocale)
		if err != nil {
			return err
		}
		p.uiLocale = l
	}
	if params.Theme != nil {
		t, err := NewTheme(*params.Theme)
		if err != nil {
			return err
		}
		p.theme = t
	}
	if params.PlaybackQuality != nil {
		q, err := NewPlaybackQuality(*params.PlaybackQuality)
		if err != nil {
			return err
		}
		p.playbackQuality = q
	}
	if params.AutoplayNext != nil {
		p.autoplayNext = *params.AutoplayNext
	}
	if params.SubtitleLanguage != nil {
		l, err := shared.NewLocale(*params.SubtitleLanguage)
		if err != nil {
			return err
		}
		p.subtitleLanguage = l
	}
	if params.SubtitlesEnabled != nil {
		p.subtitlesEnabled = *params.SubtitlesEnabled
	}
	if params.SkipIntro != nil {
		p.skipIntro = *params.SkipIntro
	}
	if params.ReducedMotion != nil {
		p.reducedMotion = *params.ReducedMotion
	}
	if params.EmailDigestFrequency != nil {
		f, err := NewDigestFrequency(*params.EmailDigestFrequency)
		if err != nil {
			return err
		}
		p.emailDigestFrequency = f
	}
	if params.NotificationSettings != nil {
		p.notificationSettings = *params.NotificationSettings
	}

	p.updatedAt = time.Now()
	return nil
}

func UnmarshalPreferences(
	accID account.AccountID,
	uiLocale string,
	theme string,
	playbackQuality string,
	autoplayNext bool,
	subtitleLanguage string,
	subtitlesEnabled bool,
	skipIntro bool,
	reducedMotion bool,
	emailDigestFrequency string,
	notificationSettings NotificationSettings,
	createdAt, updatedAt time.Time,
	version int,
) (*Preferences, error) {
	l, err := shared.NewLocale(uiLocale)
	if err != nil {
		return nil, err
	}

	t, err := NewTheme(theme)
	if err != nil {
		return nil, err
	}

	q, err := NewPlaybackQuality(playbackQuality)
	if err != nil {
		return nil, err
	}

	sl, err := shared.NewLocale(subtitleLanguage)
	if err != nil {
		return nil, err
	}

	f, err := NewDigestFrequency(emailDigestFrequency)
	if err != nil {
		return nil, err
	}

	return &Preferences{
		accountID:            accID,
		uiLocale:             l,
		theme:                t,
		playbackQuality:      q,
		autoplayNext:         autoplayNext,
		subtitleLanguage:     sl,
		subtitlesEnabled:     subtitlesEnabled,
		skipIntro:            skipIntro,
		reducedMotion:        reducedMotion,
		emailDigestFrequency: f,
		notificationSettings: notificationSettings,
		createdAt:            createdAt,
		updatedAt:            updatedAt,
		version:              version,
	}, nil
}
