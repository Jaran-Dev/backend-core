-- Categories Table
CREATE TABLE categories(
    id uuid PRIMARY KEY DEFAULT uuid_generate_v4(),
    parent_id uuid REFERENCES categories(id) ON DELETE RESTRICT,
    slug text UNIQUE NOT NULL,
    status text NOT NULL DEFAULT 'draft' CHECK (status IN ('draft', 'published', 'archived')),
    cover_url text,
    position int NOT NULL DEFAULT 0,
    deleted_at timestamptz,
    created_at timestamptz NOT NULL DEFAULT now(),
    updated_at timestamptz NOT NULL DEFAULT now(),
    version int NOT NULL DEFAULT 1
);

CREATE INDEX idx_categories_parent_id ON categories(parent_id);

CREATE INDEX idx_categories_deleted_at ON categories(deleted_at);

-- Translations Table
CREATE TABLE category_translations(
    category_id uuid NOT NULL REFERENCES categories(id) ON DELETE CASCADE,
    locale text NOT NULL,
    title text NOT NULL,
    description text,
    PRIMARY KEY (category_id, locale)
);

-- Closure Table for Hierarchy (Max Depth 2)
CREATE TABLE category_closure(
    ancestor_id uuid NOT NULL REFERENCES categories(id) ON DELETE CASCADE,
    descendant_id uuid NOT NULL REFERENCES categories(id) ON DELETE CASCADE,
    depth int NOT NULL,
    PRIMARY KEY (ancestor_id, descendant_id)
);

-- Trigger Function to maintain closure table, max depth, and prevent cycles
CREATE OR REPLACE FUNCTION maintain_category_closure()
    RETURNS TRIGGER
    AS $$
DECLARE
    v_max_depth int := 2;
    v_parent_depth int;
    v_subtree_max_depth int;
BEGIN
    -- 1. Ensure self-reference exists
    INSERT INTO category_closure(ancestor_id, descendant_id, depth)
        VALUES (NEW.id, NEW.id, 0)
    ON CONFLICT (ancestor_id, descendant_id)
        DO NOTHING;
    IF (TG_OP = 'INSERT') OR (TG_OP = 'UPDATE' AND OLD.parent_id IS NULL AND NEW.parent_id IS NOT NULL) THEN
        -- INSERT or transitioning from NULL to a parent
        IF NEW.parent_id IS NOT NULL THEN
            -- Check depth constraint
            SELECT
                COALESCE(MAX(depth), -1)
            INTO
                v_parent_depth
            FROM
                category_closure
            WHERE
                descendant_id = NEW.parent_id;
            IF v_parent_depth + 1 > v_max_depth THEN
                RAISE EXCEPTION 'Depth constraint violated: max depth is %', v_max_depth;
            END IF;
            -- Insert ancestors for the new node
            INSERT INTO category_closure(ancestor_id, descendant_id, depth)
            SELECT
                ancestor_id,
                NEW.id,
                depth + 1
            FROM
                category_closure
            WHERE
                descendant_id = NEW.parent_id;
        END IF;
    ELSIF (TG_OP = 'UPDATE'
            AND OLD.parent_id IS NOT NULL
            AND NEW.parent_id IS NOT NULL
            AND OLD.parent_id <> NEW.parent_id) THEN
        -- RE-PARENTING
        -- Cycle check
        IF EXISTS (
            SELECT
                1
            FROM
                category_closure
            WHERE
                ancestor_id = NEW.id
                AND descendant_id = NEW.parent_id) THEN
        RAISE EXCEPTION 'Circular reference detected';
    END IF;
    -- Check depth constraint for the entire subtree being moved
    SELECT
        COALESCE(MAX(depth), -1)
    INTO
        v_parent_depth
    FROM
        category_closure
    WHERE
        descendant_id = NEW.parent_id;
    SELECT
        COALESCE(MAX(depth), 0)
    INTO
        v_subtree_max_depth
    FROM
        category_closure
    WHERE
        ancestor_id = NEW.id;
    IF v_parent_depth + v_subtree_max_depth > v_max_depth THEN
        RAISE EXCEPTION 'Depth constraint violated: max depth is %', v_max_depth;
    END IF;
    -- Delete old paths (paths from old ancestors to the subtree)
    DELETE FROM category_closure
    WHERE descendant_id IN (
            SELECT
                descendant_id
            FROM
                category_closure
            WHERE
                ancestor_id = NEW.id)
        AND ancestor_id IN (
            SELECT
                ancestor_id
            FROM
                category_closure
            WHERE
                descendant_id = NEW.id
                AND ancestor_id <> NEW.id);
    -- Insert new paths
    INSERT INTO category_closure(ancestor_id, descendant_id, depth)
    SELECT
        supertree.ancestor_id,
        subtree.descendant_id,
        supertree.depth + subtree.depth + 1
    FROM
        category_closure supertree
        CROSS JOIN category_closure subtree
    WHERE
        supertree.descendant_id = NEW.parent_id
        AND subtree.ancestor_id = NEW.id;
ELSIF (TG_OP = 'UPDATE'
        AND OLD.parent_id IS NOT NULL
        AND NEW.parent_id IS NULL) THEN
    -- Detaching from parent (becoming root)
    DELETE FROM category_closure
    WHERE descendant_id IN (
            SELECT
                descendant_id
            FROM
                category_closure
            WHERE
                ancestor_id = NEW.id)
        AND ancestor_id IN (
            SELECT
                ancestor_id
            FROM
                category_closure
            WHERE
                descendant_id = NEW.id
                AND ancestor_id <> NEW.id);
END IF;
    RETURN NEW;
END;
$$
LANGUAGE plpgsql;

CREATE TRIGGER trigger_maintain_category_closure
    AFTER INSERT OR UPDATE OF parent_id ON categories
    FOR EACH ROW
    EXECUTE FUNCTION maintain_category_closure();

