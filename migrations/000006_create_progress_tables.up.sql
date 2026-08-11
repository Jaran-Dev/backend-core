CREATE TABLE enrollments(
    id uuid PRIMARY KEY DEFAULT uuid_generate_v4(),
    account_id uuid NOT NULL REFERENCES accounts(id) ON DELETE CASCADE,
    course_id uuid NOT NULL REFERENCES courses(id) ON DELETE CASCADE,
    enrolled_at timestamptz NOT NULL DEFAULT now(),
    UNIQUE (account_id, course_id)
);

CREATE TABLE lesson_progress(
    id uuid PRIMARY KEY DEFAULT uuid_generate_v4(),
    account_id uuid NOT NULL REFERENCES accounts(id) ON DELETE CASCADE,
    lesson_id uuid NOT NULL REFERENCES lessons(id) ON DELETE CASCADE,
    watched_sec int NOT NULL DEFAULT 0,
    completed boolean NOT NULL DEFAULT FALSE,
    updated_at timestamptz NOT NULL DEFAULT now(),
    UNIQUE (account_id, lesson_id)
);

CREATE INDEX idx_progress_account_lesson ON lesson_progress(account_id, lesson_id);

