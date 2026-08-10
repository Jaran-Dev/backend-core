CREATE TABLE modules (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    course_id UUID NOT NULL REFERENCES courses(id) ON DELETE CASCADE,
    position INT NOT NULL,
    title TEXT NOT NULL DEFAULT '', -- Fallback title, canonical titles live in Lessons
    UNIQUE(course_id, position)
);

CREATE INDEX idx_modules_course_id ON modules(course_id);