package out

import (
	"context"

	"github.com/google/uuid"
	"github.com/tuannm99/judge-loop/internal/domain"
)

type TestCaseRepository interface {
	GetByProblem(ctx context.Context, problemID uuid.UUID) ([]domain.TestCase, error)
}
