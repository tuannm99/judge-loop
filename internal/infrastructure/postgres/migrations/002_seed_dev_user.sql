-- +goose Up
-- Seed the development user required by the no-auth MVP.
-- The UUID matches the USER_ID / JUDGE_USER_ID default in all service configs.
-- Running this migration ensures FK constraints from submissions, sessions, and
-- training tables are satisfied on a fresh database.

INSERT INTO users (id, username, email)
VALUES ('00000000-0000-0000-0000-000000000001', 'dev', 'dev@judge-loop.local')
ON CONFLICT (id) DO NOTHING;

-- +goose Down
DELETE FROM users
WHERE id = '00000000-0000-0000-0000-000000000001';
