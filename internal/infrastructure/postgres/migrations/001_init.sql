-- +goose Up
-- Migration 001: Initial schema for judge-loop
-- PostgreSQL 15+

-- Enable pgcrypto for gen_random_uuid()
CREATE EXTENSION IF NOT EXISTS pgcrypto;

-- ============================================================
-- USERS
-- ============================================================

CREATE TABLE users (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    username    TEXT NOT NULL UNIQUE,
    email       TEXT NOT NULL UNIQUE,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- ============================================================
-- PROBLEMS
-- ============================================================

CREATE TYPE difficulty AS ENUM ('easy', 'medium', 'hard');
CREATE TYPE provider   AS ENUM ('leetcode', 'neetcode', 'hackerrank');

CREATE TABLE problems (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    slug            TEXT NOT NULL UNIQUE,
    title           TEXT NOT NULL,
    difficulty      difficulty NOT NULL,
    tags            TEXT[]    NOT NULL DEFAULT '{}',
    pattern_tags    TEXT[]    NOT NULL DEFAULT '{}',
    provider        provider  NOT NULL,
    external_id     TEXT      NOT NULL,
    source_url      TEXT      NOT NULL DEFAULT '',
    estimated_time  INT       NOT NULL DEFAULT 0,  -- minutes
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE (provider, external_id)
);

CREATE INDEX idx_problems_difficulty   ON problems (difficulty);
CREATE INDEX idx_problems_provider     ON problems (provider);
CREATE INDEX idx_problems_tags         ON problems USING GIN (tags);
CREATE INDEX idx_problems_pattern_tags ON problems USING GIN (pattern_tags);

-- ============================================================
-- TEST CASES
-- ============================================================

CREATE TABLE test_cases (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    problem_id  UUID NOT NULL REFERENCES problems (id) ON DELETE CASCADE,
    input       TEXT NOT NULL,
    expected    TEXT NOT NULL,
    is_hidden   BOOLEAN NOT NULL DEFAULT FALSE,
    order_idx   INT     NOT NULL DEFAULT 0
);

CREATE INDEX idx_test_cases_problem ON test_cases (problem_id);

-- ============================================================
-- SUBMISSIONS
-- ============================================================

CREATE TYPE submission_status AS ENUM (
    'pending',
    'running',
    'accepted',
    'wrong_answer',
    'compile_error',
    'runtime_error',
    'time_limit_exceeded'
);

CREATE TYPE language AS ENUM ('python', 'go');

CREATE TABLE submissions (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id         UUID NOT NULL REFERENCES users (id),
    problem_id      UUID NOT NULL REFERENCES problems (id),
    session_id      UUID,  -- FK to timer_sessions, nullable
    language        language NOT NULL,
    code            TEXT NOT NULL,
    status          submission_status NOT NULL DEFAULT 'pending',
    verdict         TEXT NOT NULL DEFAULT '',
    passed_cases    INT  NOT NULL DEFAULT 0,
    total_cases     INT  NOT NULL DEFAULT 0,
    runtime_ms      BIGINT NOT NULL DEFAULT 0,
    error_message   TEXT NOT NULL DEFAULT '',
    submitted_at    TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    evaluated_at    TIMESTAMPTZ
);

CREATE INDEX idx_submissions_user       ON submissions (user_id);
CREATE INDEX idx_submissions_problem    ON submissions (problem_id);
CREATE INDEX idx_submissions_status     ON submissions (status);
CREATE INDEX idx_submissions_submitted  ON submissions (submitted_at DESC);

-- ============================================================
-- TIMER SESSIONS
-- ============================================================

CREATE TABLE timer_sessions (
    id           UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id      UUID NOT NULL REFERENCES users (id),
    problem_id   UUID REFERENCES problems (id),
    started_at   TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    ended_at     TIMESTAMPTZ,
    elapsed_secs INT NOT NULL DEFAULT 0
);

CREATE INDEX idx_timer_sessions_user    ON timer_sessions (user_id);
CREATE INDEX idx_timer_sessions_started ON timer_sessions (started_at DESC);

-- Add FK from submissions to timer_sessions now that both tables exist
ALTER TABLE submissions
    ADD CONSTRAINT fk_submissions_session
    FOREIGN KEY (session_id) REFERENCES timer_sessions (id);

-- ============================================================
-- DAILY SESSIONS
-- ============================================================

CREATE TABLE daily_sessions (
    id               UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id          UUID NOT NULL REFERENCES users (id),
    date             DATE NOT NULL,
    solved_count     INT  NOT NULL DEFAULT 0,
    attempted_count  INT  NOT NULL DEFAULT 0,
    time_spent_secs  INT  NOT NULL DEFAULT 0,
    created_at       TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at       TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE (user_id, date)
);

CREATE INDEX idx_daily_sessions_user ON daily_sessions (user_id);
CREATE INDEX idx_daily_sessions_date ON daily_sessions (date DESC);

-- ============================================================
-- ACTIVITY EVENTS
-- ============================================================

CREATE TABLE activity_events (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id     UUID NOT NULL REFERENCES users (id),
    event_type  TEXT NOT NULL,
    problem_id  UUID REFERENCES problems (id),
    session_id  UUID REFERENCES timer_sessions (id),
    payload     JSONB NOT NULL DEFAULT '{}',
    occurred_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_activity_events_user     ON activity_events (user_id);
CREATE INDEX idx_activity_events_occurred ON activity_events (occurred_at DESC);

-- ============================================================
-- REVIEW SCHEDULES (spaced repetition)
-- ============================================================

CREATE TABLE review_schedules (
    id             UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id        UUID NOT NULL REFERENCES users (id),
    problem_id     UUID NOT NULL REFERENCES problems (id),
    next_review_at TIMESTAMPTZ NOT NULL,
    interval_days  INT NOT NULL DEFAULT 1,
    review_count   INT NOT NULL DEFAULT 0,
    created_at     TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at     TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE (user_id, problem_id)
);

CREATE INDEX idx_review_schedules_user       ON review_schedules (user_id);
CREATE INDEX idx_review_schedules_next_review ON review_schedules (next_review_at);

-- ============================================================
-- USER TRAINING PROFILES
-- ============================================================

CREATE TABLE user_training_profiles (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id         UUID NOT NULL UNIQUE REFERENCES users (id),
    goals           TEXT[]  NOT NULL DEFAULT '{}',
    minutes_per_day INT     NOT NULL DEFAULT 60,
    easy_pct        INT     NOT NULL DEFAULT 20,
    medium_pct      INT     NOT NULL DEFAULT 60,
    hard_pct        INT     NOT NULL DEFAULT 20,
    weak_patterns   TEXT[]  NOT NULL DEFAULT '{}',
    focus_patterns  TEXT[]  NOT NULL DEFAULT '{}',
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- ============================================================
-- TRAINING CONTRACTS
-- ============================================================

CREATE TABLE training_contracts (
    id               UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id          UUID NOT NULL REFERENCES users (id),
    daily_problems   INT  NOT NULL DEFAULT 2,
    weekly_problems  INT  NOT NULL DEFAULT 10,
    focus_time       INT  NOT NULL DEFAULT 60,  -- minutes
    review_enabled   BOOLEAN NOT NULL DEFAULT TRUE,
    active_from      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    created_at       TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_training_contracts_user ON training_contracts (user_id);

-- ============================================================
-- DAILY MISSIONS
-- ============================================================

CREATE TABLE daily_missions (
    id           UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id      UUID NOT NULL REFERENCES users (id),
    date         DATE NOT NULL,
    -- tasks stored as JSONB arrays of MissionTask
    required_tasks  JSONB NOT NULL DEFAULT '[]',
    optional_tasks  JSONB NOT NULL DEFAULT '[]',
    review_tasks    JSONB NOT NULL DEFAULT '[]',
    generated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE (user_id, date)
);

CREATE INDEX idx_daily_missions_user ON daily_missions (user_id);
CREATE INDEX idx_daily_missions_date ON daily_missions (date DESC);

-- ============================================================
-- PROBLEM PERFORMANCE
-- ============================================================

CREATE TABLE problem_performances (
    id                UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id           UUID NOT NULL REFERENCES users (id),
    problem_id        UUID NOT NULL REFERENCES problems (id),
    first_solve_time  FLOAT,   -- minutes, NULL if unsolved
    best_solve_time   FLOAT,
    latest_solve_time FLOAT,
    attempts          INT     NOT NULL DEFAULT 0,
    accepted          BOOLEAN NOT NULL DEFAULT FALSE,
    complexity        TEXT    NOT NULL DEFAULT '',  -- self-reported
    confidence        INT     NOT NULL DEFAULT 0,   -- 1-5 self-reported
    last_attempt_at   TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at        TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE (user_id, problem_id)
);

CREATE INDEX idx_problem_performances_user    ON problem_performances (user_id);
CREATE INDEX idx_problem_performances_problem ON problem_performances (problem_id);

-- ============================================================
-- PROBLEM BANK ITEMS
-- ============================================================

CREATE TABLE problem_bank_items (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id     UUID NOT NULL REFERENCES users (id),
    problem_id  UUID NOT NULL REFERENCES problems (id),
    imported    BOOLEAN NOT NULL DEFAULT FALSE,
    in_progress BOOLEAN NOT NULL DEFAULT FALSE,
    pinned      BOOLEAN NOT NULL DEFAULT FALSE,
    added_at    TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE (user_id, problem_id)
);

CREATE INDEX idx_problem_bank_user ON problem_bank_items (user_id);

-- ============================================================
-- PERFORMANCE SNAPSHOTS
-- ============================================================

CREATE TABLE performance_snapshots (
    id               UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id          UUID NOT NULL REFERENCES users (id),
    snapshot_date    DATE NOT NULL,
    avg_solve_time   FLOAT   NOT NULL DEFAULT 0,
    total_attempts   INT     NOT NULL DEFAULT 0,
    accepted_count   INT     NOT NULL DEFAULT 0,
    hint_usage_rate  FLOAT   NOT NULL DEFAULT 0,
    pattern_scores   JSONB   NOT NULL DEFAULT '{}',
    created_at       TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE (user_id, snapshot_date)
);

CREATE INDEX idx_performance_snapshots_user ON performance_snapshots (user_id);
CREATE INDEX idx_performance_snapshots_date ON performance_snapshots (snapshot_date DESC);

-- ============================================================
-- REGISTRY VERSIONS
-- ============================================================

CREATE TABLE registry_versions (
    id          SERIAL PRIMARY KEY,
    version     TEXT NOT NULL,
    updated_at  TIMESTAMPTZ NOT NULL,
    manifests   JSONB NOT NULL DEFAULT '[]',
    synced_at   TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- +goose Down
DROP TABLE IF EXISTS registry_versions;
DROP TABLE IF EXISTS performance_snapshots;
DROP TABLE IF EXISTS problem_bank_items;
DROP TABLE IF EXISTS problem_performances;
DROP TABLE IF EXISTS daily_missions;
DROP TABLE IF EXISTS training_contracts;
DROP TABLE IF EXISTS user_training_profiles;
DROP TABLE IF EXISTS review_schedules;
DROP TABLE IF EXISTS activity_events;
DROP TABLE IF EXISTS daily_sessions;
ALTER TABLE IF EXISTS submissions DROP CONSTRAINT IF EXISTS fk_submissions_session;
DROP TABLE IF EXISTS timer_sessions;
DROP TABLE IF EXISTS submissions;
DROP TABLE IF EXISTS test_cases;
DROP TABLE IF EXISTS problems;
DROP TABLE IF EXISTS users;
DROP TYPE IF EXISTS language;
DROP TYPE IF EXISTS submission_status;
DROP TYPE IF EXISTS provider;
DROP TYPE IF EXISTS difficulty;
