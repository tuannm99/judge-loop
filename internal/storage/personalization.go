package storage

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/tuannm99/judge-loop/internal/domain"
)

// ─────────────────────────────────────────
// MissionStore
// ─────────────────────────────────────────

// MissionStore handles daily mission persistence.
type MissionStore struct{ db *DB }

// NewMissionStore creates a new MissionStore.
func NewMissionStore(db *DB) *MissionStore { return &MissionStore{db: db} }

// GetToday returns today's cached mission for the user, or nil if none exists yet.
func (s *MissionStore) GetToday(ctx context.Context, userID uuid.UUID) (*domain.DailyMission, error) {
	row := s.db.Pool.QueryRow(ctx, `
		SELECT id, user_id, date, required_tasks, optional_tasks, review_tasks, generated_at
		FROM daily_missions
		WHERE user_id = $1 AND date = CURRENT_DATE`,
		userID,
	)
	return scanMission(row)
}

// Save inserts or replaces today's mission for the user.
func (s *MissionStore) Save(ctx context.Context, m domain.DailyMission) error {
	req, err := json.Marshal(m.RequiredTasks)
	if err != nil {
		return fmt.Errorf("marshal required_tasks: %w", err)
	}
	opt, err := json.Marshal(m.OptionalTasks)
	if err != nil {
		return fmt.Errorf("marshal optional_tasks: %w", err)
	}
	rev, err := json.Marshal(m.ReviewTasks)
	if err != nil {
		return fmt.Errorf("marshal review_tasks: %w", err)
	}

	_, err = s.db.Pool.Exec(ctx, `
		INSERT INTO daily_missions (id, user_id, date, required_tasks, optional_tasks, review_tasks, generated_at)
		VALUES ($1, $2, CURRENT_DATE, $3, $4, $5, $6)
		ON CONFLICT (user_id, date) DO UPDATE SET
			required_tasks = EXCLUDED.required_tasks,
			optional_tasks = EXCLUDED.optional_tasks,
			review_tasks   = EXCLUDED.review_tasks,
			generated_at   = EXCLUDED.generated_at`,
		m.ID, m.UserID, req, opt, rev, m.GeneratedAt,
	)
	if err != nil {
		return fmt.Errorf("save mission: %w", err)
	}
	return nil
}

func scanMission(row pgx.Row) (*domain.DailyMission, error) {
	var m domain.DailyMission
	var reqRaw, optRaw, revRaw []byte
	err := row.Scan(&m.ID, &m.UserID, &m.Date, &reqRaw, &optRaw, &revRaw, &m.GeneratedAt)
	if err == pgx.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("scan mission: %w", err)
	}
	_ = json.Unmarshal(reqRaw, &m.RequiredTasks)
	_ = json.Unmarshal(optRaw, &m.OptionalTasks)
	_ = json.Unmarshal(revRaw, &m.ReviewTasks)
	return &m, nil
}

// ─────────────────────────────────────────
// PerformanceStore
// ─────────────────────────────────────────

// PerformanceStore computes user performance metrics from submissions.
type PerformanceStore struct{ db *DB }

// NewPerformanceStore creates a new PerformanceStore.
func NewPerformanceStore(db *DB) *PerformanceStore { return &PerformanceStore{db: db} }

// PatternScoreRow holds per-pattern accepted/attempted counts.
type PatternScoreRow struct {
	Pattern  string
	Accepted int
	Attempted int
}

// GetPatternScores returns a map of pattern tag → score (accepted/attempted).
// Patterns with fewer than 1 submission are excluded.
func (s *PerformanceStore) GetPatternScores(ctx context.Context, userID uuid.UUID) (map[string]float64, error) {
	rows, err := s.db.Pool.Query(ctx, `
		SELECT
			unnest(p.pattern_tags)                              AS pattern,
			COUNT(*) FILTER (WHERE s.status = 'accepted')      AS accepted,
			COUNT(*)                                            AS attempted
		FROM submissions s
		JOIN problems p ON p.id = s.problem_id
		WHERE s.user_id = $1
		GROUP BY pattern
		ORDER BY pattern`,
		userID,
	)
	if err != nil {
		return nil, fmt.Errorf("pattern scores: %w", err)
	}
	defer rows.Close()

	scores := make(map[string]float64)
	for rows.Next() {
		var r PatternScoreRow
		if err := rows.Scan(&r.Pattern, &r.Accepted, &r.Attempted); err != nil {
			return nil, fmt.Errorf("scan pattern score: %w", err)
		}
		if r.Attempted > 0 {
			scores[r.Pattern] = float64(r.Accepted) / float64(r.Attempted)
		}
	}
	return scores, rows.Err()
}

// PerformanceStats holds aggregate submission statistics for a user.
type PerformanceStats struct {
	TotalAttempts int
	AcceptedCount int
	AvgSolveTime  float64 // minutes (elapsed from timer_sessions where available)
}

// GetStats returns aggregate performance statistics for the user.
func (s *PerformanceStore) GetStats(ctx context.Context, userID uuid.UUID) (PerformanceStats, error) {
	var st PerformanceStats
	err := s.db.Pool.QueryRow(ctx, `
		SELECT
			COUNT(*)                                        AS total_attempts,
			COUNT(*) FILTER (WHERE s.status = 'accepted')  AS accepted_count,
			COALESCE(AVG(
				CASE WHEN t.elapsed_secs > 0
				     THEN t.elapsed_secs / 60.0 END
			), 0)                                           AS avg_solve_minutes
		FROM submissions s
		LEFT JOIN timer_sessions t ON t.id = s.session_id
		WHERE s.user_id = $1`,
		userID,
	).Scan(&st.TotalAttempts, &st.AcceptedCount, &st.AvgSolveTime)
	if err != nil {
		return PerformanceStats{}, fmt.Errorf("get stats: %w", err)
	}
	return st, nil
}

// ─────────────────────────────────────────
// GetUnsolved (added to ProblemStore)
// ─────────────────────────────────────────

// GetUnsolved returns problems the user has not yet accepted, in random order.
func (s *ProblemStore) GetUnsolved(ctx context.Context, userID uuid.UUID, limit int) ([]domain.Problem, error) {
	if limit <= 0 {
		limit = 10
	}
	rows, err := s.db.Pool.Query(ctx, `
		SELECT`+problemCols+`
		FROM problems p
		WHERE p.id NOT IN (
			SELECT DISTINCT problem_id
			FROM submissions
			WHERE user_id = $1 AND status = 'accepted'
		)
		ORDER BY RANDOM()
		LIMIT $2`,
		userID, limit,
	)
	if err != nil {
		return nil, fmt.Errorf("get unsolved: %w", err)
	}
	defer rows.Close()

	var out []domain.Problem
	for rows.Next() {
		p, err := scanProblem(rows)
		if err != nil {
			return nil, fmt.Errorf("scan unsolved problem: %w", err)
		}
		out = append(out, p)
	}
	return out, rows.Err()
}

// ─────────────────────────────────────────
// ReviewStore.Upsert (spaced repetition)
// ─────────────────────────────────────────

// UpsertReview creates or advances the review schedule for a problem.
// Intervals (MVP): 1 → 3 → 7 → 14 days.
func (s *ReviewStore) Upsert(ctx context.Context, userID, problemID uuid.UUID) error {
	_, err := s.db.Pool.Exec(ctx, `
		INSERT INTO review_schedules (id, user_id, problem_id, next_review_at, interval_days, review_count)
		VALUES (gen_random_uuid(), $1, $2, NOW() + INTERVAL '1 day', 1, 1)
		ON CONFLICT (user_id, problem_id) DO UPDATE SET
			interval_days  = CASE
				WHEN review_schedules.review_count = 1 THEN 3
				WHEN review_schedules.review_count = 2 THEN 7
				WHEN review_schedules.review_count >= 3 THEN 14
				ELSE 1
			END,
			next_review_at = NOW() + (CASE
				WHEN review_schedules.review_count = 1 THEN INTERVAL '3 days'
				WHEN review_schedules.review_count = 2 THEN INTERVAL '7 days'
				WHEN review_schedules.review_count >= 3 THEN INTERVAL '14 days'
				ELSE INTERVAL '1 day'
			END),
			review_count   = review_schedules.review_count + 1,
			updated_at     = NOW()`,
		userID, problemID,
	)
	if err != nil {
		return fmt.Errorf("upsert review schedule: %w", err)
	}
	return nil
}

// LastSyncedAt returns when the registry was last synced, or zero time.
func (s *RegistryStore) LastSyncedAt(ctx context.Context) time.Time {
	row, _ := s.GetLatest(ctx)
	if row == nil {
		return time.Time{}
	}
	return row.SyncedAt
}
