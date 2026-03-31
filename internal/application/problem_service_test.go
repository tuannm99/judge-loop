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
	service := NewProblemService(problems)

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

func TestProblemServiceGetProblemFallsBackToSlug(t *testing.T) {
	problems := outmocks.NewMockProblemRepository(t)
	service := NewProblemService(problems)

	want := &domain.Problem{Slug: "two-sum"}
	problems.EXPECT().GetBySlug(mock.Anything, "two-sum").Return(want, nil).Once()

	got, err := service.GetProblem(context.Background(), "two-sum")
	require.NoError(t, err)
	assert.Same(t, want, got)
}

func TestProblemServiceGetProblemByID(t *testing.T) {
	problems := outmocks.NewMockProblemRepository(t)
	service := NewProblemService(problems)

	ctx := context.Background()
	problemID := uuid.New()
	want := &domain.Problem{ID: problemID}
	problems.EXPECT().GetByID(ctx, problemID).Return(want, nil)

	got, err := service.GetProblem(ctx, problemID.String())
	require.NoError(t, err)
	require.Equal(t, want, got)
}

func TestProblemServiceSuggestProblem(t *testing.T) {
	problems := outmocks.NewMockProblemRepository(t)
	service := NewProblemService(problems)

	ctx := context.Background()
	userID := uuid.New()
	problem := &domain.Problem{ID: uuid.New()}
	problems.EXPECT().Suggest(ctx, userID, []string(nil)).Return(problem, nil)

	got, err := service.SuggestProblem(ctx, userID)
	require.NoError(t, err)
	require.Equal(t, problem, got)
}
