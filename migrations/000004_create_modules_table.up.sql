CREATE TABLE modules(
    id uuid PRIMARY KEY DEFAULT uuid_generate_v4(),
    course_id uuid NOT NULL REFERENCES courses(id) ON DELETE CASCADE,
    position int NOT NULL,
    title text NOT NULL DEFAULT '',
    UNIQUE (course_id, position)
);

CREATE INDEX idx_modules_course_id ON modules(course_id);

