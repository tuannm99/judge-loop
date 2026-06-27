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

func TestSubmissionRepositoryCreateAndGet(t *testing.T) {
	db, mock := newMockDB(t)
	repository := NewSubmissionRepositoryImpl(db)
	userID := uuid.New()
	problemID := uuid.New()
	submission := &domain.Submission{
		UserID:    userID,
		ProblemID: problemID,
		Language:  domain.LanguageGo,
		Code:      "package main",
	}

	mock.ExpectQuery(`INSERT INTO "submissions".*RETURNING "id"`).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(uuid.New()))
	require.NoError(t, repository.Create(context.Background(), submission))
	require.NotEqual(t, uuid.Nil, submission.ID)
	require.Equal(t, domain.StatusPending, submission.Status)

	mock.ExpectQuery(`SELECT \* FROM "submissions" WHERE id = \$1 ORDER BY "submissions"\."id" LIMIT \$2`).
		WithArgs(submission.ID, 1).
		WillReturnRows(sqlmock.NewRows([]string{
			"id", "user_id", "problem_id", "language", "code", "status", "submitted_at",
		}).AddRow(
			submission.ID,
			userID,
			problemID,
			string(domain.LanguageGo),
			"package main",
			string(domain.StatusPending),
			time.Now(),
		))

	stored, err := repository.GetByID(context.Background(), submission.ID)
	require.NoError(t, err)
	require.Equal(t, submission.ID, stored.ID)
	require.Equal(t, domain.LanguageGo, stored.Language)
}

func TestSubmissionRepositoryTryStartAndCountSolved(t *testing.T) {
	db, mock := newMockDB(t)
	repository := NewSubmissionRepositoryImpl(db)
	submissionID := uuid.New()
	userID := uuid.New()

	mock.ExpectExec(`UPDATE "submissions" SET "status"=\$1 WHERE id = \$2 AND status = \$3`).
		WithArgs(string(domain.StatusRunning), submissionID, string(domain.StatusPending)).
		WillReturnResult(sqlmock.NewResult(0, 1))

	started, err := repository.TryStartEvaluation(context.Background(), submissionID)
	require.NoError(t, err)
	require.True(t, started)

	mock.ExpectQuery(
		`SELECT COUNT\(DISTINCT\("problem_id"\)\) FROM "submissions" WHERE user_id = \$1 AND status = \$2`,
	).
		WithArgs(userID, string(domain.StatusAccepted)).
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(3))

	count, err := repository.GetDistinctSolvedCount(context.Background(), userID)
	require.NoError(t, err)
	require.Equal(t, 3, count)
}

func TestSubmissionRepositoryUpdateVerdictAndList(t *testing.T) {
	db, mock := newMockDB(t)
	repository := NewSubmissionRepositoryImpl(db)
	submissionID := uuid.New()
	userID := uuid.New()
	problemID := uuid.New()
	evaluatedAt := time.Now().UTC()

	mock.ExpectExec(`UPDATE "submissions" SET .* WHERE id = \$[0-9]+`).
		WillReturnResult(sqlmock.NewResult(0, 1))
	require.NoError(t, repository.UpdateVerdict(
		context.Background(),
		submissionID,
		string(domain.StatusAccepted),
		string(domain.VerdictAccepted),
		2,
		2,
		8,
		"",
		&evaluatedAt,
	))

	mock.ExpectQuery(
		`SELECT \* FROM "submissions" WHERE user_id = \$1 AND problem_id = \$2 `+
			`ORDER BY submitted_at DESC LIMIT \$3 OFFSET \$4`,
	).
		WithArgs(userID, problemID, 20, 5).
		WillReturnRows(sqlmock.NewRows([]string{
			"id", "user_id", "problem_id", "status", "verdict", "submitted_at",
		}).AddRow(
			submissionID,
			userID,
			problemID,
			string(domain.StatusAccepted),
			string(domain.VerdictAccepted),
			evaluatedAt,
		))

	submissions, err := repository.ListByUser(context.Background(), userID, &problemID, 0, 5)
	require.NoError(t, err)
	require.Len(t, submissions, 1)
	require.Equal(t, domain.VerdictAccepted, submissions[0].Verdict)
}
