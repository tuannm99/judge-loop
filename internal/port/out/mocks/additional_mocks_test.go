package mocks

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"github.com/tuannm99/judge-loop/internal/domain"
)

func TestAdditionalRepositoryMocks(t *testing.T) {
	ctx := context.Background()
	userID := uuid.New()
	problemID := uuid.New()
	problems := []domain.Problem{{ID: problemID, Slug: "two-sum"}}
	mission := domain.DailyMission{ID: uuid.New(), UserID: userID}

	problemRepo := NewMockProblemRepository(t)
	require.NotNil(t, problemRepo.EXPECT())
	problemRepo.EXPECT().GetUnsolved(mock.Anything, userID, 3).
		Run(func(context.Context, uuid.UUID, int) {}).
		Return(problems, nil)
	got, err := problemRepo.GetUnsolved(ctx, userID, 3)
	require.NoError(t, err)
	require.Len(t, got, 1)
	problemRepo.EXPECT().GetUnsolved(mock.Anything, userID, 3).
		RunAndReturn(func(context.Context, uuid.UUID, int) ([]domain.Problem, error) {
			return problems, nil
		})
	_, err = problemRepo.GetUnsolved(ctx, userID, 3)
	require.NoError(t, err)

	missionRepo := NewMockMissionRepository(t)
	require.NotNil(t, missionRepo.EXPECT())
	missionRepo.EXPECT().GetToday(mock.Anything, userID).Run(func(context.Context, uuid.UUID) {}).Return(&mission, nil)
	gotMission, err := missionRepo.GetToday(ctx, userID)
	require.NoError(t, err)
	require.Equal(t, mission.ID, gotMission.ID)
	missionRepo.EXPECT().GetToday(mock.Anything, userID).RunAndReturn(func(context.Context, uuid.UUID) (*domain.DailyMission, error) {
		return &mission, nil
	})
	_, err = missionRepo.GetToday(ctx, userID)
	require.NoError(t, err)
	missionRepo.EXPECT().Save(mock.Anything, mission).Run(func(context.Context, domain.DailyMission) {}).Return(nil)
	require.NoError(t, missionRepo.Save(ctx, mission))
	missionRepo.EXPECT().Save(mock.Anything, mission).RunAndReturn(func(context.Context, domain.DailyMission) error {
		return nil
	})
	require.NoError(t, missionRepo.Save(ctx, mission))

	perfRepo := NewMockPerformanceRepository(t)
	require.NotNil(t, perfRepo.EXPECT())
	scores := map[string]float64{"dp": 0.8}
	perfRepo.EXPECT().GetPatternScores(mock.Anything, userID).Run(func(context.Context, uuid.UUID) {}).Return(scores, nil)
	gotScores, err := perfRepo.GetPatternScores(ctx, userID)
	require.NoError(t, err)
	require.Equal(t, 0.8, gotScores["dp"])
	perfRepo.EXPECT().GetPatternScores(mock.Anything, userID).RunAndReturn(func(context.Context, uuid.UUID) (map[string]float64, error) {
		return scores, nil
	})
	_, err = perfRepo.GetPatternScores(ctx, userID)
	require.NoError(t, err)

	registryRepo := NewMockRegistryRepository(t)
	require.NotNil(t, registryRepo.EXPECT())
	now := time.Now()
	registryRepo.EXPECT().Save(mock.Anything, "v2", now, []domain.ManifestRef(nil)).
		Run(func(context.Context, string, time.Time, []domain.ManifestRef) {}).
		Return(nil)
	require.NoError(t, registryRepo.Save(ctx, "v2", now, nil))
}
