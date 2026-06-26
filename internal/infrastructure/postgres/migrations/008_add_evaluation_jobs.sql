-- +goose Up
CREATE TYPE evaluation_job_status AS ENUM (
    'pending',
    'running',
    'done',
    'failed'
);

CREATE TABLE evaluation_jobs (
    id            UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    submission_id UUID NOT NULL REFERENCES submissions (id) ON DELETE CASCADE,
    user_id       UUID NOT NULL REFERENCES users (id),
    status        evaluation_job_status NOT NULL DEFAULT 'pending',
    attempts      INT NOT NULL DEFAULT 0,
    max_attempts  INT NOT NULL DEFAULT 3,
    available_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    locked_at     TIMESTAMPTZ,
    locked_by     TEXT NOT NULL DEFAULT '',
    last_error    TEXT NOT NULL DEFAULT '',
    created_at    TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at    TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE (submission_id)
);

CREATE INDEX idx_evaluation_jobs_pick
    ON evaluation_jobs (status, available_at, created_at);

CREATE INDEX idx_evaluation_jobs_submission
    ON evaluation_jobs (submission_id);

-- +goose Down
DROP TABLE IF EXISTS evaluation_jobs;
DROP TYPE IF EXISTS evaluation_job_status;
