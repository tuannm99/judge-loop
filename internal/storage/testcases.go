package storage

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/tuannm99/judge-loop/internal/domain"
)

// TestCaseStore handles test case queries.
type TestCaseStore struct{ db *DB }

// NewTestCaseStore creates a new TestCaseStore.
func NewTestCaseStore(db *DB) *TestCaseStore { return &TestCaseStore{db: db} }

// GetByProblem returns all non-hidden test cases for a problem, ordered by order_idx.
func (s *TestCaseStore) GetByProblem(ctx context.Context, problemID uuid.UUID) ([]domain.TestCase, error) {
	rows, err := s.db.Pool.Query(ctx, `
		SELECT id, problem_id, input, expected, is_hidden, order_idx
		FROM test_cases
		WHERE problem_id = $1 AND is_hidden = false
		ORDER BY order_idx`,
		problemID,
	)
	if err != nil {
		return nil, fmt.Errorf("get test cases: %w", err)
	}
	defer rows.Close()

	var out []domain.TestCase
	for rows.Next() {
		var tc domain.TestCase
		if err := rows.Scan(&tc.ID, &tc.ProblemID, &tc.Input, &tc.Expected, &tc.IsHidden, &tc.OrderIdx); err != nil {
			return nil, fmt.Errorf("scan test case: %w", err)
		}
		out = append(out, tc)
	}
	return out, rows.Err()
}
