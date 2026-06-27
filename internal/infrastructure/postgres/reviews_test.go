package postgres

import (
	"context"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

func TestReviewRepositoryGetDueAndReset(t *testing.T) {
	db, mock := newMockDB(t)
	repository := NewReviewRepositoryImpl(db)
	userID := uuid.New()
	problemID := uuid.New()
	lastSolved := time.Now().Add(-48 * time.Hour)

	mock.ExpectQuery(`(?s)SELECT.*FROM review_schedules r.*WHERE r\.user_id = \$1`).
		WithArgs(userID).
		WillReturnRows(sqlmock.NewRows([]string{
			"problem_id", "slug", "title", "last_solved", "days_overdue",
		}).AddRow(problemID, "two-sum", "Two Sum", lastSolved, 2))

	due, err := repository.GetDue(context.Background(), userID)
	require.NoError(t, err)
	require.Len(t, due, 1)
	require.Equal(t, 2, due[0].DaysOverdue)

	mock.ExpectExec(
		`UPDATE "review_schedules" SET .* WHERE user_id = \$[0-9]+ AND problem_id = \$[0-9]+`,
	).
		WillReturnResult(sqlmock.NewResult(0, 1))
	require.NoError(t, repository.Reset(context.Background(), userID, problemID))
}

func TestReviewRepositoryUpsertCreatesMissingSchedule(t *testing.T) {
	db, mock := newMockDB(t)
	repository := NewReviewRepositoryImpl(db)
	userID := uuid.New()
	problemID := uuid.New()

	mock.ExpectBegin()
	mock.ExpectQuery(
		`SELECT \* FROM "review_schedules" WHERE user_id = \$1 AND problem_id = \$2 LIMIT \$3`,
	).
		WithArgs(userID, problemID, 1).
		WillReturnRows(sqlmock.NewRows([]string{"id"}))
	mock.ExpectQuery(`INSERT INTO "review_schedules".*RETURNING "id"`).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(uuid.New()))
	mock.ExpectCommit()

	require.NoError(t, repository.Upsert(context.Background(), userID, problemID))
}

func TestReviewRepositoryUpsertAdvancesExistingSchedule(t *testing.T) {
	db, mock := newMockDB(t)
	repository := NewReviewRepositoryImpl(db)
	userID := uuid.New()
	problemID := uuid.New()

	mock.ExpectBegin()
	mock.ExpectQuery(
		`SELECT \* FROM "review_schedules" WHERE user_id = \$1 AND problem_id = \$2 LIMIT \$3`,
	).
		WithArgs(userID, problemID, 1).
		WillReturnRows(sqlmock.NewRows([]string{
			"id", "user_id", "problem_id", "review_count", "interval_days",
		}).AddRow(uuid.New(), userID, problemID, 1, 1))
	mock.ExpectExec(
		`UPDATE "review_schedules" SET .* WHERE user_id = \$[0-9]+ AND problem_id = \$[0-9]+`,
	).
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectCommit()

	require.NoError(t, repository.Upsert(context.Background(), userID, problemID))
}
