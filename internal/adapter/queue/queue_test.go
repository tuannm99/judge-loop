package queueadapter

import (
	"context"
	"errors"
	"sync/atomic"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	outport "github.com/tuannm99/judge-loop/internal/port/out"
	outmocks "github.com/tuannm99/judge-loop/internal/port/out/mocks"
)

func TestEvaluatorProcessJob(t *testing.T) {
	subID := uuid.New()
	userID := uuid.New()
	called := false

	e := NewEvaluator(2, evaluationServiceFunc(
		func(ctx context.Context, submissionID, runnerUserID uuid.UUID, timeLimitSecs int) error {
			called = true
			require.Equal(t, subID, submissionID)
			require.Equal(t, userID, runnerUserID)
			require.Equal(t, 2, timeLimitSecs)
			return nil
		},
	))
	require.NoError(t, e.ProcessJob(context.Background(), outport.EvaluationJob{
		ID:           uuid.New(),
		SubmissionID: subID,
		UserID:       userID,
	}))
	require.True(t, called)

	e.service = evaluationServiceFunc(
		func(context.Context, uuid.UUID, uuid.UUID, int) error { return errors.New("boom") },
	)
	err := e.ProcessJob(context.Background(), outport.EvaluationJob{
		ID:           uuid.New(),
		SubmissionID: subID,
		UserID:       userID,
	})
	require.Error(t, err)
	require.Contains(t, err.Error(), "evaluate submission")
}

func TestWorkerCompletesAndFailsJobs(t *testing.T) {
	successJob := outport.EvaluationJob{ID: uuid.New(), SubmissionID: uuid.New(), UserID: uuid.New()}
	failJob := outport.EvaluationJob{ID: uuid.New(), SubmissionID: uuid.New(), UserID: uuid.New()}
	jobs := []outport.EvaluationJob{successJob, failJob}
	queue := outmocks.NewMockEvaluationJobQueue(t)
	completed := make(chan uuid.UUID, 1)
	failed := make(chan uuid.UUID, 1)

	queue.EXPECT().
		ClaimEvaluationJob(mock.Anything, "test-worker").
		RunAndReturn(func(context.Context, string) (*outport.EvaluationJob, error) {
			if len(jobs) == 0 {
				return nil, nil
			}
			job := jobs[0]
			jobs = jobs[1:]
			return &job, nil
		})
	queue.EXPECT().
		CompleteEvaluationJob(mock.Anything, successJob.ID).
		RunAndReturn(func(_ context.Context, id uuid.UUID) error {
			completed <- id
			return nil
		}).
		Once()
	queue.EXPECT().
		FailEvaluationJob(mock.Anything, failJob.ID, "evaluate submission "+failJob.SubmissionID.String()+": boom").
		RunAndReturn(func(_ context.Context, id uuid.UUID, _ string) error {
			failed <- id
			return nil
		}).
		Once()

	evaluator := NewEvaluator(2, evaluationServiceFunc(
		func(_ context.Context, submissionID, _ uuid.UUID, _ int) error {
			if submissionID == failJob.SubmissionID {
				return errors.New("boom")
			}
			return nil
		},
	))
	worker := NewWorker(queue, evaluator, "test-worker", 1)

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	go worker.Run(ctx)

	require.Equal(t, successJob.ID, receiveJobID(t, completed))
	require.Equal(t, failJob.ID, receiveJobID(t, failed))
	cancel()
}

func TestNewWorkerDefaults(t *testing.T) {
	queue := outmocks.NewMockEvaluationJobQueue(t)
	worker := NewWorker(queue, NewEvaluator(1, evaluationServiceFunc(
		func(context.Context, uuid.UUID, uuid.UUID, int) error { return nil },
	)), "", 0)
	require.Equal(t, "judge-worker", worker.workerID)
	require.Equal(t, 1, worker.concurrency)
}

func TestWorkerContinuesAfterClaimError(t *testing.T) {
	queue := outmocks.NewMockEvaluationJobQueue(t)
	var claims atomic.Int32
	queue.EXPECT().
		ClaimEvaluationJob(mock.Anything, "test-worker").
		RunAndReturn(func(context.Context, string) (*outport.EvaluationJob, error) {
			claims.Add(1)
			return nil, errors.New("down")
		})

	worker := NewWorker(queue, NewEvaluator(1, evaluationServiceFunc(
		func(context.Context, uuid.UUID, uuid.UUID, int) error { return nil },
	)), "test-worker", 1)
	worker.pollInterval = time.Millisecond

	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Millisecond)
	defer cancel()
	worker.Run(ctx)

	require.Positive(t, claims.Load())
}

type evaluationServiceFunc func(ctx context.Context, submissionID, userID uuid.UUID, timeLimitSecs int) error

func (f evaluationServiceFunc) EvaluateSubmission(
	ctx context.Context,
	submissionID, userID uuid.UUID,
	timeLimitSecs int,
) error {
	return f(ctx, submissionID, userID, timeLimitSecs)
}

func receiveJobID(t *testing.T, ch <-chan uuid.UUID) uuid.UUID {
	t.Helper()
	select {
	case id := <-ch:
		return id
	case <-time.After(time.Second):
		t.Fatal("timed out waiting for job status update")
		return uuid.Nil
	}
}
