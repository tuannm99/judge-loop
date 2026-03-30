package personalization

import (
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
	"github.com/tuannm99/judge-loop/internal/domain"
	outport "github.com/tuannm99/judge-loop/internal/port/out"
)

func TestGenerate(t *testing.T) {
	userID := uuid.New()
	weakID := uuid.New()
	normalID := uuid.New()
	optionalID := uuid.New()
	reviewID := uuid.New()

	mission := Generate(userID, Input{
		PatternScores: map[string]float64{"dp": 0.2, "array": 0.9},
		Unsolved: []domain.Problem{
			{ID: weakID, Slug: "weak", Title: "Weak", Difficulty: domain.DifficultyEasy, PatternTags: []string{"dp"}},
			{ID: normalID, Slug: "normal", Title: "Normal", Difficulty: domain.DifficultyMedium, PatternTags: []string{"array"}},
			{ID: optionalID, Slug: "opt", Title: "Optional", Difficulty: domain.DifficultyHard, PatternTags: []string{"graph"}},
		},
		DueReviews: []outport.DueReview{{ProblemID: reviewID, Slug: "review", Title: "Review", DaysOverdue: 2}},
	})

	require.Equal(t, userID, mission.UserID)
	require.Len(t, mission.ReviewTasks, 1)
	require.Equal(t, "overdue by 2 day(s)", mission.ReviewTasks[0].Reason)
	require.Len(t, mission.RequiredTasks, 2)
	require.Equal(t, weakID, mission.RequiredTasks[0].ProblemID)
	require.Contains(t, mission.RequiredTasks[0].Reason, "weak pattern: dp")
	require.Equal(t, normalID, mission.RequiredTasks[1].ProblemID)
	require.Len(t, mission.OptionalTasks, 1)
	require.Equal(t, optionalID, mission.OptionalTasks[0].ProblemID)
}

func TestWeakPatternHelpers(t *testing.T) {
	weak := WeakPatterns(map[string]float64{"dp": 0.2, "graph": 0.7})
	require.Equal(t, []string{"dp"}, weak)

	require.Equal(t, "dp", firstMatch([]string{"array", "dp"}, []string{"dp"}))
	require.Empty(t, firstMatch([]string{"array"}, []string{"dp"}))
}

func TestGenerateUsesDefaultGoal(t *testing.T) {
	userID := uuid.New()
	problems := []domain.Problem{
		{ID: uuid.New(), Slug: "a", Title: "A"},
		{ID: uuid.New(), Slug: "b", Title: "B"},
		{ID: uuid.New(), Slug: "c", Title: "C"},
		{ID: uuid.New(), Slug: "d", Title: "D"},
	}

	mission := Generate(userID, Input{Unsolved: problems})
	require.Len(t, mission.RequiredTasks, 2)
	require.Len(t, mission.OptionalTasks, 2)
}

func TestGenerateWithExplicitGoalAndDueReviewReason(t *testing.T) {
	userID := uuid.New()
	problemID := uuid.New()

	mission := Generate(userID, Input{
		DailyGoal: 1,
		Unsolved: []domain.Problem{
			{ID: problemID, Slug: "one", Title: "One", PatternTags: []string{"array"}},
			{ID: uuid.New(), Slug: "two", Title: "Two"},
		},
		DueReviews: []outport.DueReview{{ProblemID: uuid.New(), Slug: "review", Title: "Review", DaysOverdue: 0}},
	})

	require.Len(t, mission.RequiredTasks, 1)
	require.Equal(t, problemID, mission.RequiredTasks[0].ProblemID)
	require.Len(t, mission.ReviewTasks, 1)
	require.Equal(t, "due for review", mission.ReviewTasks[0].Reason)
}
