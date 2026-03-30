package queueadapter

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/hibiken/asynq"
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
	postgres "github.com/tuannm99/judge-loop/internal/infrastructure/postgres"
	q "github.com/tuannm99/judge-loop/internal/infrastructure/queue"
	outport "github.com/tuannm99/judge-loop/internal/port/out"
)

func TestNewEvaluationPublisherAndPublish(t *testing.T) {
	publisher := NewEvaluationPublisher(asynq.NewClient(asynq.RedisClientOpt{Addr: "127.0.0.1:1"}))
	require.NotNil(t, publisher)

	err := publisher.PublishEvaluation(outport.EvaluateSubmissionJob{SubmissionID: uuid.NewString(), UserID: uuid.NewString()})
	require.Error(t, err)
}

func TestEvaluationPublisherSuccess(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
	defer cancel()

	container, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: testcontainers.ContainerRequest{
			Image:        "redis:7-alpine",
			ExposedPorts: []string{"6379/tcp"},
			WaitingFor:   wait.ForListeningPort("6379/tcp"),
		},
		Started: true,
	})
	if err != nil {
		t.Skipf("docker/testcontainers unavailable: %v", err)
	}
	t.Cleanup(func() {
		_ = container.Terminate(context.Background())
	})

	host, err := container.Host(ctx)
	require.NoError(t, err)
	port, err := container.MappedPort(ctx, "6379/tcp")
	require.NoError(t, err)

	client := asynq.NewClient(asynq.RedisClientOpt{Addr: host + ":" + port.Port()})
	t.Cleanup(func() { _ = client.Close() })

	publisher := NewEvaluationPublisher(client)
	require.NoError(t, publisher.PublishEvaluation(outport.EvaluateSubmissionJob{
		SubmissionID: uuid.NewString(),
		UserID:       uuid.NewString(),
	}))
}

func TestNewEvaluatorAndProcessTask(t *testing.T) {
	e := NewEvaluator(2, &postgres.DB{})
	require.NotNil(t, e)

	err := e.ProcessTask(context.Background(), asynq.NewTask(q.TypeEvaluateSubmission, []byte("{")))
	require.Error(t, err)
	require.Contains(t, err.Error(), "unmarshal payload")

	payload := q.EvaluatePayload{SubmissionID: "bad", UserID: uuid.NewString()}
	task, err := q.NewEvaluateTask(payload)
	require.NoError(t, err)
	err = e.ProcessTask(context.Background(), task)
	require.Error(t, err)
	require.Contains(t, err.Error(), "parse submission_id")

	payload = q.EvaluatePayload{SubmissionID: uuid.NewString(), UserID: "bad"}
	task, err = q.NewEvaluateTask(payload)
	require.NoError(t, err)
	err = e.ProcessTask(context.Background(), task)
	require.Error(t, err)
	require.Contains(t, err.Error(), "parse user_id")

	subID := uuid.New()
	userID := uuid.New()
	task, err = q.NewEvaluateTask(q.EvaluatePayload{SubmissionID: subID.String(), UserID: userID.String()})
	require.NoError(t, err)

	called := false
	e.service = evaluationServiceFunc(
		func(ctx context.Context, submissionID, runnerUserID uuid.UUID, timeLimitSecs int) error {
			called = true
			require.Equal(t, subID, submissionID)
			require.Equal(t, userID, runnerUserID)
			require.Equal(t, 2, timeLimitSecs)
			return nil
		},
	)
	require.NoError(t, e.ProcessTask(context.Background(), task))
	require.True(t, called)

	e.service = evaluationServiceFunc(
		func(context.Context, uuid.UUID, uuid.UUID, int) error { return errors.New("boom") },
	)
	err = e.ProcessTask(context.Background(), task)
	require.Error(t, err)
}

type evaluationServiceFunc func(ctx context.Context, submissionID, userID uuid.UUID, timeLimitSecs int) error

func (f evaluationServiceFunc) EvaluateSubmission(
	ctx context.Context,
	submissionID, userID uuid.UUID,
	timeLimitSecs int,
) error {
	return f(ctx, submissionID, userID, timeLimitSecs)
}
