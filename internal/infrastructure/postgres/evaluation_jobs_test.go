package postgres

import (
	"context"
	"errors"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
	outport "github.com/tuannm99/judge-loop/internal/port/out"
)

func TestEvaluationJobRepositoryRejectsInvalidIdentifiers(t *testing.T) {
	db, _ := newMockDB(t)
	repository := NewEvaluationJobRepositoryImpl(db)

	err := repository.PublishEvaluation(outport.EvaluateSubmissionJob{
		SubmissionID: "invalid",
		UserID:       "invalid",
	})

	require.ErrorContains(t, err, "parse submission_id")
}

func TestEvaluationJobRepositoryClaimReturnsNilWhenQueueIsEmpty(t *testing.T) {
	db, mock := newMockDB(t)
	repository := NewEvaluationJobRepositoryImpl(db)

	mock.ExpectQuery(`(?s)UPDATE evaluation_jobs.*FOR UPDATE SKIP LOCKED.*RETURNING \*`).
		WithArgs(
			evaluationJobStatusRunning,
			"worker-a",
			evaluationJobStatusPending,
			evaluationJobStatusRunning,
			int(evaluationJobVisibilityTimeout.Seconds()),
		).
		WillReturnRows(sqlmock.NewRows([]string{"id"}))

	job, err := repository.ClaimEvaluationJob(context.Background(), "worker-a")

	require.NoError(t, err)
	require.Nil(t, job)
}

func TestEvaluationJobRepositoryClaimWrapsQueryError(t *testing.T) {
	db, mock := newMockDB(t)
	repository := NewEvaluationJobRepositoryImpl(db)

	mock.ExpectQuery(`(?s)UPDATE evaluation_jobs.*RETURNING \*`).
		WillReturnError(errors.New("database unavailable"))

	job, err := repository.ClaimEvaluationJob(context.Background(), "worker-a")

	require.Nil(t, job)
	require.ErrorContains(t, err, "claim evaluation job: database unavailable")
}

func TestEvaluationJobRepositoryCompleteAndMissingFailure(t *testing.T) {
	db, mock := newMockDB(t)
	repository := NewEvaluationJobRepositoryImpl(db)
	jobID := uuid.New()

	mock.ExpectExec(`UPDATE "evaluation_jobs" SET .* WHERE id = \$[0-9]+`).
		WillReturnResult(sqlmock.NewResult(0, 1))
	require.NoError(t, repository.CompleteEvaluationJob(context.Background(), jobID))

	mock.ExpectQuery(`SELECT \* FROM "evaluation_jobs" WHERE id = \$1 ORDER BY "evaluation_jobs"\."id" LIMIT \$2`).
		WithArgs(jobID, 1).
		WillReturnRows(sqlmock.NewRows([]string{"id"}))
	require.NoError(t, repository.FailEvaluationJob(context.Background(), jobID, "boom"))
}

func TestEvaluationJobRepositoryPublishAndRetry(t *testing.T) {
	db, mock := newMockDB(t)
	repository := NewEvaluationJobRepositoryImpl(db)
	jobID := uuid.New()
	submissionID := uuid.New()
	userID := uuid.New()

	mock.ExpectQuery(`INSERT INTO "evaluation_jobs".*ON CONFLICT.*RETURNING "id"`).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(jobID))
	require.NoError(t, repository.PublishEvaluation(outport.EvaluateSubmissionJob{
		SubmissionID: submissionID.String(),
		UserID:       userID.String(),
	}))

	mock.ExpectQuery(`SELECT \* FROM "evaluation_jobs" WHERE id = \$1 ORDER BY "evaluation_jobs"\."id" LIMIT \$2`).
		WithArgs(jobID, 1).
		WillReturnRows(sqlmock.NewRows([]string{
			"id", "submission_id", "user_id", "attempts", "max_attempts",
		}).AddRow(jobID, submissionID, userID, 1, 3))
	mock.ExpectExec(`UPDATE "evaluation_jobs" SET .* WHERE id = \$[0-9]+`).
		WillReturnResult(sqlmock.NewResult(0, 1))
	require.NoError(t, repository.FailEvaluationJob(context.Background(), jobID, "retry"))
}
