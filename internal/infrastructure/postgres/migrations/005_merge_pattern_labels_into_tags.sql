-- +goose Up

CREATE TABLE problem_labels_next (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    slug        TEXT NOT NULL UNIQUE,
    name        TEXT NOT NULL,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

INSERT INTO problem_labels_next (slug, name, created_at, updated_at)
SELECT
    slug,
    MIN(name) AS name,
    MIN(created_at) AS created_at,
    MAX(updated_at) AS updated_at
FROM problem_labels
GROUP BY slug;

CREATE TABLE problem_label_links_next (
    problem_id        UUID NOT NULL REFERENCES problems (id) ON DELETE CASCADE,
    problem_label_id  UUID NOT NULL REFERENCES problem_labels_next (id) ON DELETE CASCADE,
    PRIMARY KEY (problem_id, problem_label_id)
);

INSERT INTO problem_label_links_next (problem_id, problem_label_id)
SELECT DISTINCT pll.problem_id, next_labels.id
FROM problem_label_links pll
JOIN problem_labels labels ON labels.id = pll.problem_label_id
JOIN problem_labels_next next_labels ON next_labels.slug = labels.slug;

DROP TABLE problem_label_links;
DROP TABLE problem_labels;
DROP TYPE problem_label_kind;

ALTER TABLE problem_labels_next RENAME TO problem_labels;
ALTER TABLE problem_label_links_next RENAME TO problem_label_links;

CREATE INDEX idx_problem_label_links_problem ON problem_label_links (problem_id);
CREATE INDEX idx_problem_label_links_label ON problem_label_links (problem_label_id);

-- +goose Down

CREATE TYPE problem_label_kind AS ENUM ('tag', 'pattern');

CREATE TABLE problem_labels_next (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    kind        problem_label_kind NOT NULL,
    slug        TEXT NOT NULL,
    name        TEXT NOT NULL,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE (kind, slug)
);

INSERT INTO problem_labels_next (kind, slug, name, created_at, updated_at)
SELECT
    'tag'::problem_label_kind,
    slug,
    name,
    created_at,
    updated_at
FROM problem_labels;

CREATE INDEX idx_problem_labels_kind ON problem_labels_next (kind);

CREATE TABLE problem_label_links_next (
    problem_id        UUID NOT NULL REFERENCES problems (id) ON DELETE CASCADE,
    problem_label_id  UUID NOT NULL REFERENCES problem_labels_next (id) ON DELETE CASCADE,
    PRIMARY KEY (problem_id, problem_label_id)
);

INSERT INTO problem_label_links_next (problem_id, problem_label_id)
SELECT pll.problem_id, next_labels.id
FROM problem_label_links pll
JOIN problem_labels labels ON labels.id = pll.problem_label_id
JOIN problem_labels_next next_labels
  ON next_labels.kind = 'tag'::problem_label_kind
 AND next_labels.slug = labels.slug;

DROP TABLE problem_label_links;
DROP TABLE problem_labels;

ALTER TABLE problem_labels_next RENAME TO problem_labels;
ALTER TABLE problem_label_links_next RENAME TO problem_label_links;

CREATE INDEX idx_problem_label_links_problem ON problem_label_links (problem_id);
CREATE INDEX idx_problem_label_links_label ON problem_label_links (problem_label_id);
