package in

import (
	"context"

	"github.com/google/uuid"
	"github.com/tuannm99/judge-loop/internal/domain"
	out "github.com/tuannm99/judge-loop/internal/port/out"
)

type ProblemService interface {
	ListProblems(ctx context.Context, filter out.ProblemFilter) ([]domain.Problem, int, error)
	GetProblem(ctx context.Context, rawID string) (*domain.Problem, error)
	SuggestProblem(ctx context.Context, userID uuid.UUID) (*domain.Problem, error)
}
