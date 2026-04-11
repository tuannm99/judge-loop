package application

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"github.com/tuannm99/judge-loop/internal/domain"
	outport "github.com/tuannm99/judge-loop/internal/port/out"
	outmocks "github.com/tuannm99/judge-loop/internal/port/out/mocks"
)

func TestProblemServiceListProblems(t *testing.T) {
	problems := outmocks.NewMockProblemRepository(t)
	testCases := outmocks.NewMockTestCaseRepository(t)
	service := NewProblemService(problems, testCases)

	ctx := context.Background()
	problemID := uuid.New()
	diff := domain.DifficultyEasy
	filter := outport.ProblemFilter{Difficulty: &diff, Limit: 10}
	expectedProblems := []domain.Problem{{ID: problemID}}

	problems.EXPECT().List(ctx, filter).Return(expectedProblems, 1, nil)

	gotProblems, total, err := service.ListProblems(ctx, filter)
	require.NoError(t, err)
	require.Equal(t, expectedProblems, gotProblems)
	require.Equal(t, 1, total)
}

func TestProblemServiceListProblemLabels(t *testing.T) {
	problems := outmocks.NewMockProblemRepository(t)
	testCases := outmocks.NewMockTestCaseRepository(t)
	service := NewProblemService(problems, testCases)

	ctx := context.Background()
	problems.EXPECT().ListLabels(ctx, "tag").Return([]string{"array", "graph"}, nil).Once()
	problems.EXPECT().ListLabels(ctx, "pattern").Return([]string{"dp", "greedy"}, nil).Once()

	tags, patterns, err := service.ListProblemLabels(ctx)
	require.NoError(t, err)
	require.Equal(t, []string{"array", "graph"}, tags)
	require.Equal(t, []string{"dp", "greedy"}, patterns)
}

func TestProblemServiceProblemLabelCRUD(t *testing.T) {
	problems := outmocks.NewMockProblemRepository(t)
	testCases := outmocks.NewMockTestCaseRepository(t)
	service := NewProblemService(problems, testCases)

	ctx := context.Background()
	labelID := uuid.New()

	problems.EXPECT().ListLabelRecords(ctx, "tag").Return([]domain.ProblemLabel{{ID: labelID, Kind: "tag", Slug: "array", Name: "Array"}}, nil).Once()
	labels, err := service.ListProblemLabelRecords(ctx, " tag ")
	require.NoError(t, err)
	require.Len(t, labels, 1)

	problems.EXPECT().CreateLabel(ctx, domain.ProblemLabel{Kind: "tag", Slug: "graph", Name: "Graph"}).
		Return(&domain.ProblemLabel{ID: labelID, Kind: "tag", Slug: "graph", Name: "Graph"}, nil).Once()
	created, err := service.CreateProblemLabel(ctx, "tag", " graph ", "Graph")
	require.NoError(t, err)
	require.Equal(t, "graph", created.Slug)

	problems.EXPECT().UpdateLabel(ctx, domain.ProblemLabel{ID: labelID, Slug: "graphs", Name: "Graphs"}).
		Return(&domain.ProblemLabel{ID: labelID, Kind: "tag", Slug: "graphs", Name: "Graphs"}, nil).Once()
	updated, err := service.UpdateProblemLabel(ctx, labelID, " graphs ", "Graphs")
	require.NoError(t, err)
	require.Equal(t, "graphs", updated.Slug)

	problems.EXPECT().DeleteLabel(ctx, labelID).Return(nil).Once()
	require.NoError(t, service.DeleteProblemLabel(ctx, labelID))
}

func TestProblemServiceGetProblemFallsBackToSlug(t *testing.T) {
	problems := outmocks.NewMockProblemRepository(t)
	testCases := outmocks.NewMockTestCaseRepository(t)
	service := NewProblemService(problems, testCases)

	want := &domain.Problem{Slug: "two-sum"}
	problems.EXPECT().GetBySlug(mock.Anything, "two-sum").Return(want, nil).Once()

	got, err := service.GetProblem(context.Background(), "two-sum")
	require.NoError(t, err)
	assert.Same(t, want, got)
}

func TestProblemServiceGetProblemByID(t *testing.T) {
	problems := outmocks.NewMockProblemRepository(t)
	testCases := outmocks.NewMockTestCaseRepository(t)
	service := NewProblemService(problems, testCases)

	ctx := context.Background()
	problemID := uuid.New()
	want := &domain.Problem{ID: problemID}
	problems.EXPECT().GetByID(ctx, problemID).Return(want, nil)

	got, err := service.GetProblem(ctx, problemID.String())
	require.NoError(t, err)
	require.Equal(t, want, got)
}

func TestProblemServiceUpdateProblem(t *testing.T) {
	problems := outmocks.NewMockProblemRepository(t)
	testCases := outmocks.NewMockTestCaseRepository(t)
	service := NewProblemService(problems, testCases)

	ctx := context.Background()
	problemID := uuid.New()
	manifest := domain.ProblemManifest{
		Provider:      domain.ProviderLeetCode,
		ExternalID:    "1",
		Slug:          "two-sum",
		Title:         "Two Sum",
		Difficulty:    domain.DifficultyEasy,
		Tags:          []string{"array", "hash-table"},
		PatternTags:   []string{"two-pointers"},
		SourceURL:     "https://example.com/two-sum",
		EstimatedTime: 15,
		StarterCode:   map[string]string{"python": "class Solution:\n    pass\n"},
	}
	want := &domain.Problem{ID: problemID, Slug: "two-sum"}
	problems.EXPECT().Update(ctx, problemID, manifest).Return(want, nil).Once()

	got, err := service.UpdateProblem(ctx, problemID, manifest)
	require.NoError(t, err)
	require.Equal(t, want, got)
}

func TestProblemServiceUpdateProblemWithTestCases(t *testing.T) {
	problems := outmocks.NewMockProblemRepository(t)
	testCases := outmocks.NewMockTestCaseRepository(t)
	service := NewProblemService(problems, testCases)

	ctx := context.Background()
	problemID := uuid.New()
	manifest := domain.ProblemManifest{Slug: "two-sum", Title: "Two Sum"}
	saved := &domain.Problem{ID: problemID, Slug: "two-sum"}
	inputCases := []domain.TestCase{
		{Input: "1 2", Expected: "3"},
		{Input: "2 3", Expected: "5", IsHidden: true},
	}

	problems.EXPECT().Update(ctx, problemID, manifest).Return(saved, nil).Once()
	testCases.EXPECT().ReplaceForProblem(ctx, problemID, []domain.TestCase{
		{ProblemID: problemID, Input: "1 2", Expected: "3", OrderIdx: 0},
		{ProblemID: problemID, Input: "2 3", Expected: "5", IsHidden: true, OrderIdx: 1},
	}).Return(nil).Once()
	problems.EXPECT().GetByID(ctx, problemID).Return(saved, nil).Once()

	got, err := service.UpdateProblemWithTestCases(ctx, problemID, manifest, inputCases)
	require.NoError(t, err)
	require.Equal(t, saved, got)
}

func TestProblemServiceSuggestProblem(t *testing.T) {
	problems := outmocks.NewMockProblemRepository(t)
	testCases := outmocks.NewMockTestCaseRepository(t)
	service := NewProblemService(problems, testCases)

	ctx := context.Background()
	userID := uuid.New()
	problem := &domain.Problem{ID: uuid.New()}
	problems.EXPECT().Suggest(ctx, userID, []string(nil)).Return(problem, nil)

	got, err := service.SuggestProblem(ctx, userID)
	require.NoError(t, err)
	require.Equal(t, problem, got)
}

func TestProblemServiceContributeProblem(t *testing.T) {
	problems := outmocks.NewMockProblemRepository(t)
	testCases := outmocks.NewMockTestCaseRepository(t)
	service := NewProblemService(problems, testCases)

	ctx := context.Background()
	manifest := domain.ProblemManifest{
		Provider:   domain.ProviderLeetCode,
		ExternalID: "1",
		Slug:       "two-sum",
		Title:      "Two Sum",
		StarterCode: map[string]string{
			"python": "class Solution:\n    pass\n",
		},
	}
	saved := &domain.Problem{
		ID:    uuid.New(),
		Slug:  "two-sum",
		Title: "Two Sum",
		StarterCode: map[string]string{
			"python": "class Solution:\n    pass\n",
		},
	}
	inputCases := []domain.TestCase{
		{Input: "1 2", Expected: "3"},
		{Input: "2 3", Expected: "5", IsHidden: true},
	}

	problems.EXPECT().UpsertFromManifest(ctx, manifest).Return(nil).Once()
	problems.EXPECT().GetBySlug(ctx, "two-sum").Return(saved, nil).Once()
	testCases.EXPECT().ReplaceForProblem(ctx, saved.ID, []domain.TestCase{
		{ProblemID: saved.ID, Input: "1 2", Expected: "3", OrderIdx: 0},
		{ProblemID: saved.ID, Input: "2 3", Expected: "5", IsHidden: true, OrderIdx: 1},
	}).Return(nil).Once()

	got, err := service.ContributeProblem(ctx, manifest, inputCases)
	require.NoError(t, err)
	require.Equal(t, saved, got)
}
