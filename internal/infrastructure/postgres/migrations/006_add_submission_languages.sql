-- +goose Up

ALTER TYPE language ADD VALUE IF NOT EXISTS 'javascript';
ALTER TYPE language ADD VALUE IF NOT EXISTS 'typescript';
ALTER TYPE language ADD VALUE IF NOT EXISTS 'rust';

-- +goose Down

ALTER TABLE submissions
    ALTER COLUMN language TYPE TEXT
    USING language::TEXT;

DROP TYPE language;

CREATE TYPE language AS ENUM ('python', 'go');

ALTER TABLE submissions
    ALTER COLUMN language TYPE language
    USING language::language;
