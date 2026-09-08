CREATE TABLE course_categories(
    course_id uuid NOT NULL REFERENCES courses(id) ON DELETE CASCADE,
    category_id uuid NOT NULL REFERENCES categories(id) ON DELETE RESTRICT,
    is_primary boolean NOT NULL DEFAULT FALSE,
    position int NOT NULL DEFAULT 0,
    created_at timestamptz NOT NULL DEFAULT now(),
    PRIMARY KEY (course_id, category_id)
);

-- Enforce that a course can only have ONE primary category
CREATE UNIQUE INDEX idx_course_primary_category ON course_categories(course_id)
WHERE
    is_primary = TRUE;

CREATE INDEX idx_course_categories_category_id ON course_categories(category_id);

