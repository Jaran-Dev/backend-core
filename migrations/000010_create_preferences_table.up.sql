CREATE TABLE preferences(
    account_id uuid PRIMARY KEY REFERENCES accounts(id) ON DELETE CASCADE,
    ui_locale varchar(5) NOT NULL DEFAULT 'en',
    theme varchar(10) NOT NULL DEFAULT 'system',
    playback_quality varchar(10) NOT NULL DEFAULT 'auto',
    autoplay_next boolean NOT NULL DEFAULT TRUE,
    subtitle_language varchar(5) NOT NULL DEFAULT 'en',
    subtitles_enabled boolean NOT NULL DEFAULT TRUE,
    skip_intro boolean NOT NULL DEFAULT TRUE,
    reduced_motion boolean NOT NULL DEFAULT FALSE,
    email_digest_frequency varchar(10) NOT NULL DEFAULT 'weekly',
    notification_settings jsonb NOT NULL DEFAULT '{"course_progress": true, "new_lessons": true, "live_events": true, "marketing": false, "security_alerts": true, "weekly_summary": true, "product_updates": true}',
    created_at timestamptz NOT NULL DEFAULT now(),
    updated_at timestamptz NOT NULL DEFAULT now(),
    version int NOT NULL DEFAULT 1
);

