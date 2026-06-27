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

func TestSessionRepositoryGetOrCreateToday(t *testing.T) {
	db, mock := newMockDB(t)
	repository := NewSessionRepositoryImpl(db)
	userID := uuid.New()
	sessionID := uuid.New()
	today := dateOnlyUTC(time.Now())

	mock.ExpectQuery(`INSERT INTO "daily_sessions".*ON CONFLICT.*DO NOTHING RETURNING "id"`).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(sessionID))
	mock.ExpectQuery(
		`SELECT \* FROM "daily_sessions" WHERE user_id = \$1 AND date = \$2 LIMIT \$3`,
	).
		WithArgs(userID, today.Format("2006-01-02"), 1).
		WillReturnRows(sqlmock.NewRows([]string{
			"id", "user_id", "date", "solved_count", "attempted_count", "time_spent_secs",
		}).AddRow(sessionID, userID, today, 2, 3, 120))

	session, err := repository.GetOrCreateToday(context.Background(), userID)
	require.NoError(t, err)
	require.Equal(t, sessionID, session.ID)
	require.Equal(t, 3, session.AttemptedCount)
}

func TestSessionRepositoryActiveTimerReturnsNilWhenMissing(t *testing.T) {
	db, mock := newMockDB(t)
	repository := NewSessionRepositoryImpl(db)
	userID := uuid.New()

	mock.ExpectQuery(
		`SELECT \* FROM "timer_sessions" WHERE user_id = \$1 AND ended_at IS NULL ORDER BY started_at DESC LIMIT \$2`,
	).
		WithArgs(userID, 1).
		WillReturnRows(sqlmock.NewRows([]string{"id"}))

	timer, err := repository.ActiveTimer(context.Background(), userID)
	require.NoError(t, err)
	require.Nil(t, timer)
}

func TestSessionRepositoryRecordSubmission(t *testing.T) {
	db, mock := newMockDB(t)
	repository := NewSessionRepositoryImpl(db)

	mock.ExpectBegin()
	mock.ExpectQuery(`INSERT INTO "daily_sessions".*ON CONFLICT.*DO NOTHING RETURNING "id"`).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(uuid.New()))
	mock.ExpectExec(`UPDATE "daily_sessions" SET .* WHERE user_id = \$[0-9]+ AND date = \$[0-9]+`).
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectCommit()

	require.NoError(t, repository.RecordSubmission(context.Background(), uuid.New(), true))
}

func TestSessionRepositoryStartAndStopTimerPaths(t *testing.T) {
	t.Run("start timer", func(t *testing.T) {
		db, mock := newMockDB(t)
		repository := NewSessionRepositoryImpl(db)
		userID := uuid.New()
		problemID := uuid.New()

		mock.ExpectExec(`UPDATE "timer_sessions" SET .* WHERE user_id = \$[0-9]+ AND ended_at IS NULL`).
			WillReturnResult(sqlmock.NewResult(0, 1))
		mock.ExpectQuery(`INSERT INTO "timer_sessions".*RETURNING "id"`).
			WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(uuid.New()))

		timer, err := repository.StartTimer(context.Background(), userID, &problemID)
		require.NoError(t, err)
		require.Equal(t, userID, timer.UserID)
		require.Equal(t, problemID, *timer.ProblemID)
	})

	t.Run("stop timer missing", func(t *testing.T) {
		db, mock := newMockDB(t)
		repository := NewSessionRepositoryImpl(db)
		userID := uuid.New()
		mock.ExpectQuery(
			`SELECT \* FROM "timer_sessions" WHERE user_id = \$1 AND ended_at IS NULL ORDER BY started_at DESC LIMIT \$2`,
		).
			WithArgs(userID, 1).
			WillReturnRows(sqlmock.NewRows([]string{"id"}))

		timer, err := repository.StopTimer(context.Background(), userID)
		require.NoError(t, err)
		require.Nil(t, timer)
	})

	t.Run("stop timer", func(t *testing.T) {
		db, mock := newMockDB(t)
		repository := NewSessionRepositoryImpl(db)
		userID := uuid.New()
		timerID := uuid.New()
		startedAt := time.Now().Add(-time.Minute)

		mock.ExpectQuery(
			`SELECT \* FROM "timer_sessions" WHERE user_id = \$1 AND ended_at IS NULL ORDER BY started_at DESC LIMIT \$2`,
		).
			WithArgs(userID, 1).
			WillReturnRows(sqlmock.NewRows([]string{
				"id", "user_id", "started_at", "elapsed_secs",
			}).AddRow(timerID, userID, startedAt, 0))
		mock.ExpectExec(`UPDATE "timer_sessions" SET .* WHERE "id" = \$[0-9]+`).
			WillReturnResult(sqlmock.NewResult(0, 1))
		mock.ExpectBegin()
		mock.ExpectQuery(`INSERT INTO "daily_sessions".*ON CONFLICT.*DO NOTHING RETURNING "id"`).
			WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(uuid.New()))
		mock.ExpectExec(`UPDATE "daily_sessions" SET .* WHERE user_id = \$[0-9]+ AND date = \$[0-9]+`).
			WillReturnResult(sqlmock.NewResult(0, 1))
		mock.ExpectCommit()

		timer, err := repository.StopTimer(context.Background(), userID)
		require.NoError(t, err)
		require.NotNil(t, timer.EndedAt)
		require.GreaterOrEqual(t, timer.ElapsedSecs, 59)
	})
}

func TestSessionRepositoryGetStreakAndElapsed(t *testing.T) {
	db, mock := newMockDB(t)
	repository := NewSessionRepositoryImpl(db)
	userID := uuid.New()
	today := dateOnlyUTC(time.Now())

	mock.ExpectQuery(
		`SELECT "date" FROM "daily_sessions" WHERE user_id = \$1 AND solved_count > 0 ORDER BY date DESC`,
	).
		WithArgs(userID).
		WillReturnRows(sqlmock.NewRows([]string{"date"}).
			AddRow(today).
			AddRow(today.AddDate(0, 0, -1)))

	streak, err := repository.GetStreak(context.Background(), userID)
	require.NoError(t, err)
	require.Equal(t, 2, streak.Current)

	require.Zero(t, ElapsedNow(nil))
	endedAt := time.Now()
	require.Equal(t, 7, ElapsedNow(&domain.TimerSession{EndedAt: &endedAt, ElapsedSecs: 7}))
}
