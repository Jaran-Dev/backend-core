CREATE TABLE courses(
    id uuid PRIMARY KEY DEFAULT uuid_generate_v4(),
    slug text UNIQUE NOT NULL,
    status text NOT NULL DEFAULT 'draft', -- draft, published, archived
    cover_url text,
    created_by uuid NOT NULL REFERENCES accounts(id),
    created_at timestamptz NOT NULL DEFAULT now(),
    updated_at timestamptz NOT NULL DEFAULT now()
);

CREATE TABLE course_translations(
    course_id uuid NOT NULL REFERENCES courses(id) ON DELETE CASCADE,
    locale text NOT NULL, -- en, ckb, kmr
    title text NOT NULL,
    description text,
    PRIMARY KEY (course_id, locale)
);

CREATE INDEX idx_courses_slug ON courses(slug);

