-- +goose Up
ALTER TABLE problems
    ADD COLUMN IF NOT EXISTS execution_spec JSONB NOT NULL DEFAULT '{}'::jsonb,
    ADD COLUMN IF NOT EXISTS judge_ready BOOLEAN NOT NULL DEFAULT FALSE;

ALTER TABLE test_cases
    ADD COLUMN IF NOT EXISTS name TEXT NOT NULL DEFAULT '',
    ADD COLUMN IF NOT EXISTS input_json JSONB,
    ADD COLUMN IF NOT EXISTS expected_json JSONB,
    ADD COLUMN IF NOT EXISTS metadata JSONB DEFAULT '{}'::jsonb;

-- +goose Down
ALTER TABLE test_cases
    DROP COLUMN IF EXISTS metadata,
    DROP COLUMN IF EXISTS expected_json,
    DROP COLUMN IF EXISTS input_json,
    DROP COLUMN IF EXISTS name;

ALTER TABLE problems
    DROP COLUMN IF EXISTS judge_ready,
    DROP COLUMN IF EXISTS execution_spec;
