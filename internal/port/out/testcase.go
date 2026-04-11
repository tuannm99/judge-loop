package out

import (
	"context"

	"github.com/google/uuid"
	"github.com/tuannm99/judge-loop/internal/domain"
)

type TestCaseRepository interface {
	GetByProblem(ctx context.Context, problemID uuid.UUID) ([]domain.TestCase, error)
	GetAllByProblem(ctx context.Context, problemID uuid.UUID) ([]domain.TestCase, error)
	ReplaceForProblem(ctx context.Context, problemID uuid.UUID, testCases []domain.TestCase) error
}
