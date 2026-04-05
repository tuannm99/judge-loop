-- +goose Up
ALTER TABLE problems
    ADD COLUMN IF NOT EXISTS starter_code JSONB NOT NULL DEFAULT '{}'::jsonb;

-- +goose Down
ALTER TABLE problems
    DROP COLUMN IF EXISTS starter_code;
