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

const submissionCols = `
	id, user_id, problem_id, session_id, language::text, code, status::text,
	verdict, passed_cases, total_cases, runtime_ms, error_message,
	submitted_at, evaluated_at`

// SubmissionStore handles all submission queries.
type SubmissionStore struct{ db *DB }

// NewSubmissionStore creates a new SubmissionStore.
func NewSubmissionStore(db *DB) *SubmissionStore { return &SubmissionStore{db: db} }

func scanSubmission(r rowScanner) (domain.Submission, error) {
	var s domain.Submission
	var lang, status string
	err := r.Scan(
		&s.ID, &s.UserID, &s.ProblemID, &s.SessionID,
		&lang, &s.Code, &status,
		&s.Verdict, &s.PassedCases, &s.TotalCases, &s.RuntimeMS, &s.ErrorMessage,
		&s.SubmittedAt, &s.EvaluatedAt,
	)
	if err != nil {
		return domain.Submission{}, err
	}
	s.Language = domain.Language(lang)
	s.Status = domain.SubmissionStatus(status)
	return s, nil
}

// Create inserts a new submission with status=pending and returns the ID.
func (s *SubmissionStore) Create(ctx context.Context, sub *domain.Submission) error {
	sub.ID = uuid.New()
	sub.Status = domain.StatusPending
	sub.SubmittedAt = time.Now()

	_, err := s.db.Pool.Exec(ctx, `
		INSERT INTO submissions
			(id, user_id, problem_id, session_id, language, code, status, submitted_at)
		VALUES ($1, $2, $3, $4, $5::language, $6, 'pending'::submission_status, $7)`,
		sub.ID, sub.UserID, sub.ProblemID, sub.SessionID,
		string(sub.Language), sub.Code, sub.SubmittedAt,
	)
	if err != nil {
		return fmt.Errorf("create submission: %w", err)
	}
	return nil
}

// GetByID returns a submission by UUID, or nil if not found.
func (s *SubmissionStore) GetByID(ctx context.Context, id uuid.UUID) (*domain.Submission, error) {
	row := s.db.Pool.QueryRow(ctx, `SELECT`+submissionCols+`FROM submissions WHERE id = $1`, id)
	sub, err := scanSubmission(row)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("get submission: %w", err)
	}
	return &sub, nil
}

// UpdateVerdict writes the final verdict back to the submission row.
func (s *SubmissionStore) UpdateVerdict(
	ctx context.Context,
	id uuid.UUID,
	status, verdict string,
	passed, total int,
	runtimeMS int64,
	errMsg string,
	evaluatedAt *time.Time,
) error {
	_, err := s.db.Pool.Exec(ctx, `
		UPDATE submissions
		SET status       = $2::submission_status,
		    verdict      = $3,
		    passed_cases = $4,
		    total_cases  = $5,
		    runtime_ms   = $6,
		    error_message = $7,
		    evaluated_at = $8
		WHERE id = $1`,
		id, status, verdict, passed, total, runtimeMS, errMsg, evaluatedAt,
	)
	if err != nil {
		return fmt.Errorf("update verdict: %w", err)
	}
	return nil
}

// ListByUser returns the user's submission history, ordered by most recent.
func (s *SubmissionStore) ListByUser(ctx context.Context, userID uuid.UUID, problemID *uuid.UUID, limit, offset int) ([]domain.Submission, error) {
	if limit <= 0 {
		limit = 20
	}
	const q = `
		SELECT` + submissionCols + `
		FROM submissions
		WHERE user_id = $1
		  AND ($2::uuid IS NULL OR problem_id = $2)
		ORDER BY submitted_at DESC
		LIMIT $3 OFFSET $4`

	rows, err := s.db.Pool.Query(ctx, q, userID, problemID, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("list submissions: %w", err)
	}
	defer rows.Close()

	var out []domain.Submission
	for rows.Next() {
		sub, err := scanSubmission(rows)
		if err != nil {
			return nil, fmt.Errorf("scan submission: %w", err)
		}
		out = append(out, sub)
	}
	return out, rows.Err()
}
