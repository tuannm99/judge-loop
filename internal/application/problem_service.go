package application

import (
	"context"

	"github.com/google/uuid"
	"github.com/tuannm99/judge-loop/internal/domain"
	outport "github.com/tuannm99/judge-loop/internal/port/out"
)

type ProblemService struct {
	problems outport.ProblemRepository
}

func NewProblemService(problems outport.ProblemRepository) *ProblemService {
	return &ProblemService{problems: problems}
}

func (s *ProblemService) ListProblems(ctx context.Context, filter outport.ProblemFilter) ([]domain.Problem, int, error) {
	return s.problems.List(ctx, filter)
}

func (s *ProblemService) GetProblem(ctx context.Context, rawID string) (*domain.Problem, error) {
	if id, err := uuid.Parse(rawID); err == nil {
		return s.problems.GetByID(ctx, id)
	}
	return s.problems.GetBySlug(ctx, rawID)
}

func (s *ProblemService) SuggestProblem(ctx context.Context, userID uuid.UUID) (*domain.Problem, error) {
	return s.problems.Suggest(ctx, userID, nil)
}
