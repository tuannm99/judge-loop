-- +goose Up

CREATE TYPE problem_label_kind AS ENUM ('tag', 'pattern');

CREATE TABLE problem_labels (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    kind        problem_label_kind NOT NULL,
    slug        TEXT NOT NULL,
    name        TEXT NOT NULL,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE (kind, slug)
);

CREATE INDEX idx_problem_labels_kind ON problem_labels (kind);

CREATE TABLE problem_label_links (
    problem_id        UUID NOT NULL REFERENCES problems (id) ON DELETE CASCADE,
    problem_label_id  UUID NOT NULL REFERENCES problem_labels (id) ON DELETE CASCADE,
    PRIMARY KEY (problem_id, problem_label_id)
);

CREATE INDEX idx_problem_label_links_problem ON problem_label_links (problem_id);
CREATE INDEX idx_problem_label_links_label   ON problem_label_links (problem_label_id);

INSERT INTO problem_labels (kind, slug, name)
SELECT DISTINCT 'tag'::problem_label_kind, tag, tag
FROM problems, unnest(tags) AS tag
WHERE tag <> ''
ON CONFLICT (kind, slug) DO NOTHING;

INSERT INTO problem_labels (kind, slug, name)
SELECT DISTINCT 'pattern'::problem_label_kind, pattern, pattern
FROM problems, unnest(pattern_tags) AS pattern
WHERE pattern <> ''
ON CONFLICT (kind, slug) DO NOTHING;

INSERT INTO problem_label_links (problem_id, problem_label_id)
SELECT DISTINCT p.id, l.id
FROM problems p
JOIN LATERAL unnest(p.tags) AS tag(slug) ON true
JOIN problem_labels l ON l.kind = 'tag'::problem_label_kind AND l.slug = tag.slug
WHERE tag.slug <> ''
ON CONFLICT DO NOTHING;

INSERT INTO problem_label_links (problem_id, problem_label_id)
SELECT DISTINCT p.id, l.id
FROM problems p
JOIN LATERAL unnest(p.pattern_tags) AS pattern(slug) ON true
JOIN problem_labels l ON l.kind = 'pattern'::problem_label_kind AND l.slug = pattern.slug
WHERE pattern.slug <> ''
ON CONFLICT DO NOTHING;

ALTER TABLE problems
    DROP COLUMN IF EXISTS tags,
    DROP COLUMN IF EXISTS pattern_tags;

-- +goose Down

ALTER TABLE problems
    ADD COLUMN IF NOT EXISTS tags TEXT[] NOT NULL DEFAULT '{}',
    ADD COLUMN IF NOT EXISTS pattern_tags TEXT[] NOT NULL DEFAULT '{}';

UPDATE problems p
SET tags = COALESCE((
    SELECT array_agg(l.slug ORDER BY l.slug)
    FROM problem_label_links pll
    JOIN problem_labels l ON l.id = pll.problem_label_id
    WHERE pll.problem_id = p.id AND l.kind = 'tag'::problem_label_kind
), '{}');

UPDATE problems p
SET pattern_tags = COALESCE((
    SELECT array_agg(l.slug ORDER BY l.slug)
    FROM problem_label_links pll
    JOIN problem_labels l ON l.id = pll.problem_label_id
    WHERE pll.problem_id = p.id AND l.kind = 'pattern'::problem_label_kind
), '{}');

DROP TABLE IF EXISTS problem_label_links;
DROP TABLE IF EXISTS problem_labels;
DROP TYPE IF EXISTS problem_label_kind;
