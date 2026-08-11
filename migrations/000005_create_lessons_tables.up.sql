CREATE TABLE lessons(
    id uuid PRIMARY KEY DEFAULT uuid_generate_v4(),
    module_id uuid NOT NULL REFERENCES modules(id) ON DELETE CASCADE,
    position int NOT NULL,
    access_type text NOT NULL DEFAULT 'preview',
    asset_id text,
    playback_id text,
    duration_sec int NOT NULL DEFAULT 0,
    status text NOT NULL DEFAULT 'uploading',
    created_at timestamptz NOT NULL DEFAULT now(),
    updated_at timestamptz NOT NULL DEFAULT now(),
    UNIQUE (module_id, position)
);

CREATE TABLE lesson_translations(
    lesson_id uuid NOT NULL REFERENCES lessons(id) ON DELETE CASCADE,
    locale text NOT NULL,
    title text NOT NULL,
    description text,
    PRIMARY KEY (lesson_id, locale)
);

CREATE TABLE subtitles(
    id uuid PRIMARY KEY DEFAULT uuid_generate_v4(),
    lesson_id uuid NOT NULL REFERENCES lessons(id) ON DELETE CASCADE,
    locale text NOT NULL,
    url text NOT NULL,
    UNIQUE (lesson_id, locale)
);

CREATE INDEX idx_lessons_module_id ON lessons(module_id);

