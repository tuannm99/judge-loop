package postgres

import (
	"context"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
	"github.com/tuannm99/judge-loop/internal/domain"
)

func TestMissionRepositoryGetTodayMapsTasks(t *testing.T) {
	db, mock := newMockDB(t)
	repository := NewMissionRepositoryImpl(db)
	userID := uuid.New()
	missionID := uuid.New()
	now := time.Now().UTC()

	mock.ExpectQuery(
		`SELECT \* FROM "daily_missions" WHERE user_id = \$1 AND date = CURRENT_DATE LIMIT \$2`,
	).
		WithArgs(userID, 1).
		WillReturnRows(sqlmock.NewRows([]string{
			"id", "user_id", "date", "required_tasks", "optional_tasks", "review_tasks", "generated_at",
		}).AddRow(
			missionID,
			userID,
			now,
			[]byte(`[{"slug":"two-sum","title":"Two Sum"}]`),
			[]byte(`[]`),
			[]byte(`[]`),
			now,
		))

	mission, err := repository.GetToday(context.Background(), userID)
	require.NoError(t, err)
	require.Equal(t, missionID, mission.ID)
	require.Len(t, mission.RequiredTasks, 1)
	require.Equal(t, "two-sum", mission.RequiredTasks[0].Slug)
}

func TestPerformanceRepositoryMapsScoresAndStats(t *testing.T) {
	db, mock := newMockDB(t)
	repository := NewPerformanceRepositoryImpl(db)
	userID := uuid.New()

	mock.ExpectQuery(`(?s)SELECT.*AS pattern.*FROM submissions s.*WHERE s\.user_id = \$1`).
		WithArgs(userID).
		WillReturnRows(sqlmock.NewRows([]string{"pattern", "accepted", "attempted"}).
			AddRow("array", 3, 4).
			AddRow("empty", 0, 0))

	scores, err := repository.GetPatternScores(context.Background(), userID)
	require.NoError(t, err)
	require.InEpsilon(t, 0.75, scores["array"], 0.001)
	_, exists := scores["empty"]
	require.False(t, exists)

	mock.ExpectQuery(`(?s)SELECT.*AS total_attempts.*FROM submissions s.*WHERE s\.user_id = \$1`).
		WithArgs(userID).
		WillReturnRows(sqlmock.NewRows([]string{
			"total_attempts", "accepted_count", "avg_solve_minutes",
		}).AddRow(8, 5, 12.5))

	stats, err := repository.GetStats(context.Background(), userID)
	require.NoError(t, err)
	require.Equal(t, 8, stats.TotalAttempts)
	require.Equal(t, 5, stats.AcceptedCount)
	require.InEpsilon(t, 12.5, stats.AvgSolveTime, 0.001)
}

func TestMissionRepositorySaveAndRegistryLastSynced(t *testing.T) {
	db, mock := newMockDB(t)
	missionRepository := NewMissionRepositoryImpl(db)
	registryRepository := NewRegistryRepositoryImpl(db)
	now := time.Now().UTC()

	mock.ExpectQuery(`INSERT INTO "daily_missions".*ON CONFLICT.*RETURNING "id"`).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(uuid.New()))
	require.NoError(t, missionRepository.Save(context.Background(), domain.DailyMission{
		ID:          uuid.New(),
		UserID:      uuid.New(),
		Date:        now,
		GeneratedAt: now,
	}))

	mock.ExpectQuery(`SELECT \* FROM "registry_versions" ORDER BY id DESC LIMIT \$1`).
		WithArgs(1).
		WillReturnRows(sqlmock.NewRows([]string{
			"id", "version", "updated_at", "synced_at",
		}).AddRow(1, "v1", now, now))
	require.Equal(t, now, registryRepository.LastSyncedAt(context.Background()))
}
