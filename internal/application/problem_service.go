package application

import (
	"context"
	"strings"

	"github.com/google/uuid"
	"github.com/tuannm99/judge-loop/internal/domain"
	inport "github.com/tuannm99/judge-loop/internal/port/in"
	outport "github.com/tuannm99/judge-loop/internal/port/out"
)

type ProblemService struct {
	problems  outport.ProblemRepository
	testCases outport.TestCaseRepository
}

var _ inport.ProblemService = (*ProblemService)(nil)

func NewProblemService(
	problems outport.ProblemRepository,
	testCases outport.TestCaseRepository,
) *ProblemService {
	return &ProblemService{problems: problems, testCases: testCases}
}

func (s *ProblemService) ListProblems(
	ctx context.Context,
	filter outport.ProblemFilter,
) ([]domain.Problem, int, error) {
	return s.problems.List(ctx, filter)
}

func (s *ProblemService) ListProblemLabels(ctx context.Context) ([]string, []string, error) {
	tags, err := s.problems.ListLabels(ctx, "tag")
	if err != nil {
		return nil, nil, err
	}
	patterns, err := s.problems.ListLabels(ctx, "pattern")
	if err != nil {
		return nil, nil, err
	}
	return tags, patterns, nil
}

func (s *ProblemService) ListProblemLabelRecords(ctx context.Context, kind string) ([]domain.ProblemLabel, error) {
	return s.problems.ListLabelRecords(ctx, strings.TrimSpace(kind))
}

func (s *ProblemService) CreateProblemLabel(ctx context.Context, kind, slug, name string) (*domain.ProblemLabel, error) {
	label := domain.ProblemLabel{
		Kind: strings.TrimSpace(kind),
		Slug: strings.TrimSpace(slug),
		Name: strings.TrimSpace(name),
	}
	if label.Name == "" {
		label.Name = label.Slug
	}
	return s.problems.CreateLabel(ctx, label)
}

func (s *ProblemService) UpdateProblemLabel(ctx context.Context, id uuid.UUID, slug, name string) (*domain.ProblemLabel, error) {
	label := domain.ProblemLabel{
		ID:   id,
		Slug: strings.TrimSpace(slug),
		Name: strings.TrimSpace(name),
	}
	if label.Name == "" {
		label.Name = label.Slug
	}
	return s.problems.UpdateLabel(ctx, label)
}

func (s *ProblemService) DeleteProblemLabel(ctx context.Context, id uuid.UUID) error {
	return s.problems.DeleteLabel(ctx, id)
}

func (s *ProblemService) GetProblem(ctx context.Context, rawID string) (*domain.Problem, error) {
	if id, err := uuid.Parse(rawID); err == nil {
		return s.problems.GetByID(ctx, id)
	}
	return s.problems.GetBySlug(ctx, rawID)
}

func (s *ProblemService) GetProblemTestCases(ctx context.Context, id uuid.UUID) ([]domain.TestCase, error) {
	return s.testCases.GetAllByProblem(ctx, id)
}

func (s *ProblemService) UpdateProblem(
	ctx context.Context,
	id uuid.UUID,
	manifest domain.ProblemManifest,
) (*domain.Problem, error) {
	return s.problems.Update(ctx, id, manifest)
}

func (s *ProblemService) UpdateProblemWithTestCases(
	ctx context.Context,
	id uuid.UUID,
	manifest domain.ProblemManifest,
	testCases []domain.TestCase,
) (*domain.Problem, error) {
	problem, err := s.problems.Update(ctx, id, manifest)
	if err != nil || problem == nil {
		return problem, err
	}
	if err := s.testCases.ReplaceForProblem(ctx, id, normalizeTestCases(id, testCases)); err != nil {
		return nil, err
	}
	return s.problems.GetByID(ctx, id)
}

func (s *ProblemService) SuggestProblem(ctx context.Context, userID uuid.UUID) (*domain.Problem, error) {
	return s.problems.Suggest(ctx, userID, nil)
}

func (s *ProblemService) ContributeProblem(
	ctx context.Context,
	manifest domain.ProblemManifest,
	testCases []domain.TestCase,
) (*domain.Problem, error) {
	if err := s.problems.UpsertFromManifest(ctx, manifest); err != nil {
		return nil, err
	}
	problem, err := s.problems.GetBySlug(ctx, manifest.Slug)
	if err != nil || problem == nil {
		return problem, err
	}
	if err := s.testCases.ReplaceForProblem(ctx, problem.ID, normalizeTestCases(problem.ID, testCases)); err != nil {
		return nil, err
	}
	return problem, nil
}

func normalizeTestCases(problemID uuid.UUID, testCases []domain.TestCase) []domain.TestCase {
	out := make([]domain.TestCase, 0, len(testCases))
	for i, tc := range testCases {
		tc.ProblemID = problemID
		tc.OrderIdx = i
		out = append(out, tc)
	}
	return out
}
