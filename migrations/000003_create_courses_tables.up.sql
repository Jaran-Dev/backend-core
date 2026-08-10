CREATE TABLE courses (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    slug TEXT UNIQUE NOT NULL,
    status TEXT NOT NULL DEFAULT 'draft', -- draft, published, archived
    cover_url TEXT,
    created_by UUID NOT NULL REFERENCES admins(id),
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE course_translations (
    course_id UUID NOT NULL REFERENCES courses(id) ON DELETE CASCADE,
    locale TEXT NOT NULL, -- en, ckb, kmr
    title TEXT NOT NULL,
    description TEXT,
    PRIMARY KEY (course_id, locale)
);

CREATE INDEX idx_courses_slug ON courses(slug);