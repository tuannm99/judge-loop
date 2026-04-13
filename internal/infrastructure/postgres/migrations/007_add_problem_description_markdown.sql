-- +goose Up
ALTER TABLE problems
    ADD COLUMN IF NOT EXISTS description_markdown TEXT NOT NULL DEFAULT '';

-- +goose Down
ALTER TABLE problems
    DROP COLUMN IF EXISTS description_markdown;
