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
	service := NewProgressService(sessions)

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
	service := NewProgressService(sessions)

	ctx := context.Background()
	userID := uuid.New()
	streak := outport.StreakInfo{Current: 4}
	sessions.EXPECT().GetStreak(ctx, userID).Return(streak, nil)

	got, err := service.GetStreak(ctx, userID)
	require.NoError(t, err)
	require.Equal(t, streak, got)
}

func TestProgressServiceErrors(t *testing.T) {
	ctx := context.Background()
	userID := uuid.New()

	sessions := outmocks.NewMockSessionRepository(t)
	sessions.EXPECT().GetOrCreateToday(ctx, userID).Return(nil, errors.New("daily"))
	_, err := NewProgressService(sessions).GetProgressToday(ctx, userID)
	require.Error(t, err)

	sessions2 := outmocks.NewMockSessionRepository(t)
	sessions2.EXPECT().GetOrCreateToday(ctx, userID).Return(&domain.DailySession{}, nil)
	sessions2.EXPECT().GetStreak(ctx, userID).Return(outport.StreakInfo{}, errors.New("streak"))
	_, err = NewProgressService(sessions2).GetProgressToday(ctx, userID)
	require.Error(t, err)
}
