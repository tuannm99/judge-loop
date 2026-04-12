package in

import (
	"context"

	"github.com/google/uuid"
	"github.com/tuannm99/judge-loop/internal/domain"
	out "github.com/tuannm99/judge-loop/internal/port/out"
)

type ProblemService interface {
	ListProblems(ctx context.Context, filter out.ProblemFilter) ([]domain.Problem, int, error)
	ListProblemLabels(ctx context.Context) ([]string, error)
	ListProblemLabelRecords(ctx context.Context, kind string) ([]domain.ProblemLabel, error)
	CreateProblemLabel(ctx context.Context, kind, slug, name string) (*domain.ProblemLabel, error)
	UpdateProblemLabel(ctx context.Context, id uuid.UUID, slug, name string) (*domain.ProblemLabel, error)
	DeleteProblemLabel(ctx context.Context, id uuid.UUID) error
	GetProblem(ctx context.Context, rawID string) (*domain.Problem, error)
	GetProblemTestCases(ctx context.Context, id uuid.UUID) ([]domain.TestCase, error)
	UpdateProblem(ctx context.Context, id uuid.UUID, manifest domain.ProblemManifest) (*domain.Problem, error)
	SuggestProblem(ctx context.Context, userID uuid.UUID) (*domain.Problem, error)
	ContributeProblem(
		ctx context.Context,
		manifest domain.ProblemManifest,
		testCases []domain.TestCase,
	) (*domain.Problem, error)
	UpdateProblemWithTestCases(
		ctx context.Context,
		id uuid.UUID,
		manifest domain.ProblemManifest,
		testCases []domain.TestCase,
	) (*domain.Problem, error)
}
