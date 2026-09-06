package profilehttp

import (
	domainPref "github.com/Jaran-Dev/backend-core/internal/domain/preferences"
)

type ProfileResponse struct {
	AccountID   string `json:"account_id"`
	Email       string `json:"email"`
	DisplayName string `json:"display_name"`
	Username    string `json:"username"`
	Bio         string `json:"bio"`
	Headline    string `json:"headline"`
	Pronouns    string `json:"pronouns"`
	Location    string `json:"location"`
	DateOfBirth string `json:"date_of_birth,omitempty"`
	AvatarURL   string `json:"avatar_url"`
	CoverURL    string `json:"cover_url"`
}

type UpdateProfileRequest struct {
	DisplayName *string `json:"display_name" binding:"omitempty,min=1,max=50"`
	Username    *string `json:"username" binding:"omitempty,min=3,max=30"`
	Bio         *string `json:"bio" binding:"omitempty,max=500"`
	Headline    *string `json:"headline" binding:"omitempty,max=100"`
	Pronouns    *string `json:"pronouns" binding:"omitempty,max=20"`
	Location    *string `json:"location" binding:"omitempty,max=100"`
	DateOfBirth *string `json:"date_of_birth" binding:"omitempty"`
}

type PreferencesResponse struct {
	UILocale             string                          `json:"ui_locale"`
	Theme                string                          `json:"theme"`
	PlaybackQuality      string                          `json:"playback_quality"`
	AutoplayNext         bool                            `json:"autoplay_next"`
	SubtitleLanguage     string                          `json:"subtitle_language"`
	SubtitlesEnabled     bool                            `json:"subtitles_enabled"`
	SkipIntro            bool                            `json:"skip_intro"`
	ReducedMotion        bool                            `json:"reduced_motion"`
	EmailDigestFrequency string                          `json:"email_digest_frequency"`
	NotificationSettings domainPref.NotificationSettings `json:"notification_settings"`
}

type UpdatePreferencesRequest struct {
	UILocale             *string                          `json:"ui_locale" binding:"omitempty"`
	Theme                *string                          `json:"theme" binding:"omitempty,oneof=light dark system"`
	PlaybackQuality      *string                          `json:"playback_quality" binding:"omitempty,oneof=auto 1080p 720p 480p"`
	AutoplayNext         *bool                            `json:"autoplay_next"`
	SubtitleLanguage     *string                          `json:"subtitle_language" binding:"omitempty"`
	SubtitlesEnabled     *bool                            `json:"subtitles_enabled"`
	SkipIntro            *bool                            `json:"skip_intro"`
	ReducedMotion        *bool                            `json:"reduced_motion"`
	EmailDigestFrequency *string                          `json:"email_digest_frequency" binding:"omitempty,oneof=daily weekly monthly never"`
	NotificationSettings *domainPref.NotificationSettings `json:"notification_settings"`
}
