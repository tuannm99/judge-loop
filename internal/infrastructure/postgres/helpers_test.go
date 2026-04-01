package postgres

import (
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
	"github.com/tuannm99/judge-loop/internal/domain"
)

func TestComputeStreak(t *testing.T) {
	today := time.Now().UTC().Truncate(24 * time.Hour)
	yesterday := today.AddDate(0, 0, -1)
	twoDaysAgo := today.AddDate(0, 0, -2)
	fourDaysAgo := today.AddDate(0, 0, -4)

	t.Run("empty", func(t *testing.T) {
		info := computeStreak(nil)
		require.Zero(t, info.Current)
		require.Zero(t, info.Longest)
		require.Nil(t, info.LastPracticed)
	})

	t.Run("current from today", func(t *testing.T) {
		info := computeStreak([]time.Time{today, yesterday, twoDaysAgo, fourDaysAgo})
		require.Equal(t, 3, info.Current)
		require.Equal(t, 3, info.Longest)
		require.NotNil(t, info.LastPracticed)
		require.Equal(t, today, info.LastPracticed.UTC())
	})

	t.Run("current from yesterday", func(t *testing.T) {
		info := computeStreak([]time.Time{yesterday, twoDaysAgo})
		require.Equal(t, 2, info.Current)
		require.Equal(t, 2, info.Longest)
	})

	t.Run("longest without current", func(t *testing.T) {
		oldStart := today.AddDate(0, 0, -10)
		info := computeStreak([]time.Time{
			today.AddDate(0, 0, -3),
			today.AddDate(0, 0, -4),
			oldStart,
			oldStart.AddDate(0, 0, -1),
			oldStart.AddDate(0, 0, -2),
		})
		require.Zero(t, info.Current)
		require.Equal(t, 3, info.Longest)
	})
}

func TestMissionModelFromDomain(t *testing.T) {
	mission := domain.DailyMission{
		ID:     uuid.New(),
		UserID: uuid.New(),
		Date:   time.Now().UTC().Truncate(24 * time.Hour),
		RequiredTasks: []domain.MissionTask{
			{ProblemID: uuid.New(), Slug: "two-sum", Title: "Two Sum", Reason: "required", Priority: 10},
		},
		OptionalTasks: []domain.MissionTask{
			{ProblemID: uuid.New(), Slug: "reverse", Title: "Reverse", Reason: "optional", Priority: 5},
		},
		ReviewTasks: []domain.MissionTask{
			{ProblemID: uuid.New(), Slug: "dp", Title: "DP", Reason: "review", Priority: 8},
		},
		GeneratedAt: time.Now().UTC(),
	}

	model, err := missionModelFromDomain(mission)
	require.NoError(t, err)
	require.Equal(t, mission.ID, model.ID)
	require.Equal(t, mission.UserID, model.UserID)
	require.Equal(t, mission.Date, model.Date)
	require.NotEmpty(t, model.RequiredTasks)
	require.NotEmpty(t, model.OptionalTasks)
	require.NotEmpty(t, model.ReviewTasks)
	require.Equal(t, mission.GeneratedAt, model.GeneratedAt)
}
