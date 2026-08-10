CREATE TABLE lessons (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    module_id UUID NOT NULL REFERENCES modules(id) ON DELETE CASCADE,
    position INT NOT NULL,
    access_type TEXT NOT NULL DEFAULT 'preview', -- free, preview, premium
    asset_id TEXT, -- Mux/Cloudflare Asset ID
    playback_id TEXT, -- ID used to generate signed playback URLs
    duration_sec INT NOT NULL DEFAULT 0,
    status TEXT NOT NULL DEFAULT 'uploading', -- uploading, ready, error
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    UNIQUE(module_id, position)
);

CREATE TABLE lesson_translations (
    lesson_id UUID NOT NULL REFERENCES lessons(id) ON DELETE CASCADE,
    locale TEXT NOT NULL,
    title TEXT NOT NULL,
    description TEXT,
    PRIMARY KEY (lesson_id, locale)
);

CREATE TABLE subtitles (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    lesson_id UUID NOT NULL REFERENCES lessons(id) ON DELETE CASCADE,
    locale TEXT NOT NULL,
    url TEXT NOT NULL,
    UNIQUE(lesson_id, locale)
);

CREATE INDEX idx_lessons_module_id ON lessons(module_id);