package application

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
	"github.com/tuannm99/judge-loop/internal/domain"
	inport "github.com/tuannm99/judge-loop/internal/port/in"
	outport "github.com/tuannm99/judge-loop/internal/port/out"
	outmocks "github.com/tuannm99/judge-loop/internal/port/out/mocks"
)

func TestProgressServiceGetProgressToday(t *testing.T) {
	sessions := outmocks.NewMockSessionRepository(t)
	service := NewProgressService(sessions, nil)

	ctx := context.Background()
	userID := uuid.New()
	ds := &domain.DailySession{
		Date:           time.Date(2026, 3, 31, 0, 0, 0, 0, time.UTC),
		SolvedCount:    2,
		AttemptedCount: 3,
		TimeSpentSecs:  125,
	}
	streak := outport.StreakInfo{Current: 4}
	sessions.EXPECT().GetOrCreateToday(ctx, userID).Return(ds, nil)
	sessions.EXPECT().GetStreak(ctx, userID).Return(streak, nil)

	progress, err := service.GetProgressToday(ctx, userID)
	require.NoError(t, err)
	require.Equal(t, inport.ProgressToday{
		Date:             "2026-03-31",
		Solved:           2,
		Attempted:        3,
		TimeSpentMinutes: 2,
		Streak:           4,
	}, progress)
}

func TestProgressServiceGetStreak(t *testing.T) {
	sessions := outmocks.NewMockSessionRepository(t)
	service := NewProgressService(sessions, nil)

	ctx := context.Background()
	userID := uuid.New()
	streak := outport.StreakInfo{Current: 4}
	sessions.EXPECT().GetStreak(ctx, userID).Return(streak, nil)

	got, err := service.GetStreak(ctx, userID)
	require.NoError(t, err)
	require.Equal(t, streak, got)
}

func TestProgressServiceGetProgressTodayErrors(t *testing.T) {
	ctx := context.Background()
	userID := uuid.New()

	t.Run("daily session error", func(t *testing.T) {
		sessions := outmocks.NewMockSessionRepository(t)
		sessions.EXPECT().GetOrCreateToday(ctx, userID).Return(nil, errors.New("daily"))
		_, err := NewProgressService(sessions, nil).GetProgressToday(ctx, userID)
		require.Error(t, err)
	})

	t.Run("streak error", func(t *testing.T) {
		sessions := outmocks.NewMockSessionRepository(t)
		sessions.EXPECT().GetOrCreateToday(ctx, userID).Return(&domain.DailySession{}, nil)
		sessions.EXPECT().GetStreak(ctx, userID).Return(outport.StreakInfo{}, errors.New("streak"))
		_, err := NewProgressService(sessions, nil).GetProgressToday(ctx, userID)
		require.Error(t, err)
	})
}

func TestProgressServiceGetGoalProgress(t *testing.T) {
	ctx := context.Background()
	userID := uuid.New()

	progress, err := NewProgressService(nil, nil).GetGoalProgress(ctx, userID)
	require.NoError(t, err)
	require.Equal(t, inport.GoalProgress{Total: inport.GoalTotal}, progress)
}

func TestProgressServiceGetGoalProgressFromSubmissions(t *testing.T) {
	ctx := context.Background()
	userID := uuid.New()
	cases := []struct {
		name          string
		solved        int
		wantRemaining int
		wantPercent   float64
	}{
		{name: "in progress", solved: 12, wantRemaining: inport.GoalTotal - 12, wantPercent: 0.4},
		{name: "over goal", solved: inport.GoalTotal + 7, wantRemaining: 0, wantPercent: 100.23333333333333},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			submissions := outmocks.NewMockSubmissionRepository(t)
			submissions.EXPECT().GetDistinctSolvedCount(ctx, userID).Return(tc.solved, nil)
			progress, err := NewProgressService(nil, submissions).GetGoalProgress(ctx, userID)
			require.NoError(t, err)
			require.Equal(t, tc.solved, progress.Solved)
			require.Equal(t, tc.wantRemaining, progress.Remaining)
			require.InEpsilon(t, tc.wantPercent, progress.Percent, 0.001)
		})
	}
}

func TestProgressServiceGetGoalProgressReturnsCountError(t *testing.T) {
	ctx := context.Background()
	userID := uuid.New()
	submissions := outmocks.NewMockSubmissionRepository(t)
	submissions.EXPECT().GetDistinctSolvedCount(ctx, userID).Return(0, errors.New("count"))
	_, err := NewProgressService(nil, submissions).GetGoalProgress(ctx, userID)
	require.Error(t, err)
}
