package storage

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/tuannm99/judge-loop/internal/domain"
)

// SessionStore handles DailySession and TimerSession queries.
type SessionStore struct{ db *DB }

// NewSessionStore creates a new SessionStore.
func NewSessionStore(db *DB) *SessionStore { return &SessionStore{db: db} }

// GetOrCreateToday returns today's DailySession for the user, creating it if needed.
func (s *SessionStore) GetOrCreateToday(ctx context.Context, userID uuid.UUID) (*domain.DailySession, error) {
	row := s.db.Pool.QueryRow(ctx, `
		INSERT INTO daily_sessions (user_id, date)
		VALUES ($1, CURRENT_DATE)
		ON CONFLICT (user_id, date) DO UPDATE
			SET date = EXCLUDED.date
		RETURNING id, user_id, date, solved_count, attempted_count, time_spent_secs, created_at, updated_at`,
		userID,
	)
	var ds domain.DailySession
	err := row.Scan(
		&ds.ID, &ds.UserID, &ds.Date,
		&ds.SolvedCount, &ds.AttemptedCount, &ds.TimeSpentSecs,
		&ds.CreatedAt, &ds.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("get or create daily session: %w", err)
	}
	return &ds, nil
}

// RecordSubmission updates today's DailySession after a submission is evaluated.
// It always increments attempted_count; it increments solved_count only when accepted=true.
func (s *SessionStore) RecordSubmission(ctx context.Context, userID uuid.UUID, accepted bool) error {
	solved := 0
	if accepted {
		solved = 1
	}
	_, err := s.db.Pool.Exec(ctx, `
		INSERT INTO daily_sessions (user_id, date, attempted_count, solved_count)
		VALUES ($1, CURRENT_DATE, 1, $2)
		ON CONFLICT (user_id, date) DO UPDATE
			SET attempted_count = daily_sessions.attempted_count + 1,
			    solved_count    = daily_sessions.solved_count + $2,
			    updated_at      = NOW()`,
		userID, solved,
	)
	if err != nil {
		return fmt.Errorf("record submission: %w", err)
	}
	return nil
}

// StartTimer stops any existing active timer and starts a new one.
func (s *SessionStore) StartTimer(ctx context.Context, userID uuid.UUID, problemID *uuid.UUID) (*domain.TimerSession, error) {
	// auto-stop any running timer so only one is active at a time
	_, _ = s.db.Pool.Exec(ctx, `
		UPDATE timer_sessions
		SET ended_at     = NOW(),
		    elapsed_secs = GREATEST(0, EXTRACT(EPOCH FROM (NOW() - started_at))::int)
		WHERE user_id = $1 AND ended_at IS NULL`,
		userID,
	)

	id := uuid.New()
	row := s.db.Pool.QueryRow(ctx, `
		INSERT INTO timer_sessions (id, user_id, problem_id, started_at)
		VALUES ($1, $2, $3, NOW())
		RETURNING id, user_id, problem_id, started_at, ended_at, elapsed_secs`,
		id, userID, problemID,
	)
	return scanTimerRow(row)
}

// StopTimer stops the currently active timer for the user and records elapsed
// time in today's DailySession.
func (s *SessionStore) StopTimer(ctx context.Context, userID uuid.UUID) (*domain.TimerSession, error) {
	row := s.db.Pool.QueryRow(ctx, `
		UPDATE timer_sessions
		SET ended_at     = NOW(),
		    elapsed_secs = GREATEST(0, EXTRACT(EPOCH FROM (NOW() - started_at))::int)
		WHERE user_id = $1 AND ended_at IS NULL
		RETURNING id, user_id, problem_id, started_at, ended_at, elapsed_secs`,
		userID,
	)
	ts, err := scanTimerRow(row)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil // no active timer
	}
	if err != nil {
		return nil, fmt.Errorf("stop timer: %w", err)
	}

	// add elapsed time to today's session
	_, _ = s.db.Pool.Exec(ctx, `
		INSERT INTO daily_sessions (user_id, date, time_spent_secs)
		VALUES ($1, CURRENT_DATE, $2)
		ON CONFLICT (user_id, date) DO UPDATE
			SET time_spent_secs = daily_sessions.time_spent_secs + $2,
			    updated_at      = NOW()`,
		userID, ts.ElapsedSecs,
	)

	return ts, nil
}

// ActiveTimer returns the currently running timer for the user, or nil.
func (s *SessionStore) ActiveTimer(ctx context.Context, userID uuid.UUID) (*domain.TimerSession, error) {
	row := s.db.Pool.QueryRow(ctx, `
		SELECT id, user_id, problem_id, started_at, ended_at, elapsed_secs
		FROM timer_sessions
		WHERE user_id = $1 AND ended_at IS NULL
		ORDER BY started_at DESC
		LIMIT 1`,
		userID,
	)
	ts, err := scanTimerRow(row)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("active timer: %w", err)
	}
	return ts, nil
}

func scanTimerRow(r rowScanner) (*domain.TimerSession, error) {
	var ts domain.TimerSession
	err := r.Scan(
		&ts.ID, &ts.UserID, &ts.ProblemID,
		&ts.StartedAt, &ts.EndedAt, &ts.ElapsedSecs,
	)
	if err != nil {
		return nil, err
	}
	return &ts, nil
}

// ElapsedNow returns seconds elapsed for an active (not-yet-stopped) timer.
func ElapsedNow(ts *domain.TimerSession) int {
	if ts == nil {
		return 0
	}
	if ts.EndedAt != nil {
		return ts.ElapsedSecs
	}
	return int(time.Since(ts.StartedAt).Seconds())
}
