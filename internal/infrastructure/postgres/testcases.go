package postgres

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/tuannm99/judge-loop/internal/domain"
	outport "github.com/tuannm99/judge-loop/internal/port/out"
)

// TestCaseRepositoryImpl handles test case queries.
type TestCaseRepositoryImpl struct{ db *DB }

var _ outport.TestCaseRepository = (*TestCaseRepositoryImpl)(nil)

// NewTestCaseRepositoryImpl creates a new TestCaseRepositoryImpl.
func NewTestCaseRepositoryImpl(db *DB) *TestCaseRepositoryImpl { return &TestCaseRepositoryImpl{db: db} }

// GetByProblem returns all non-hidden test cases for a problem, ordered by order_idx.
func (s *TestCaseRepositoryImpl) GetByProblem(ctx context.Context, problemID uuid.UUID) ([]domain.TestCase, error) {
	var models []testCaseModel
	if err := s.db.Gorm.WithContext(ctx).
		Where("problem_id = ? AND is_hidden = false", problemID).
		Order("order_idx").
		Find(&models).Error; err != nil {
		return nil, fmt.Errorf("get test cases: %w", err)
	}

	out := make([]domain.TestCase, 0, len(models))
	for _, model := range models {
		out = append(out, model.toDomain())
	}
	return out, nil
}
