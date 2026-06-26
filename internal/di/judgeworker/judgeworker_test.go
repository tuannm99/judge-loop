package dijudge

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
	"github.com/tuannm99/judge-loop/internal/config"
	inmocks "github.com/tuannm99/judge-loop/internal/port/in/mocks"
	outmocks "github.com/tuannm99/judge-loop/internal/port/out/mocks"
	"go.uber.org/fx/fxtest"
)

func newPostgresDSN(t *testing.T) string {
	t.Helper()

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()

	container, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: testcontainers.ContainerRequest{
			Image:        "postgres:16-alpine",
			ExposedPorts: []string{"5432/tcp"},
			Env: map[string]string{
				"POSTGRES_USER":     "judge",
				"POSTGRES_PASSWORD": "judge",
				"POSTGRES_DB":       "judge",
			},
			WaitingFor: wait.ForListeningPort("5432/tcp"),
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
	port, err := container.MappedPort(ctx, "5432/tcp")
	require.NoError(t, err)

	return fmt.Sprintf(
		"host=%s port=%s user=judge password=judge dbname=judge sslmode=disable",
		host,
		port.Port(),
	)
}

func TestNew(t *testing.T) {
	require.NotNil(t, New(config.JudgeWorker{}))
}

func TestProvideDB(t *testing.T) {
	dsn := newPostgresDSN(t)

	t.Run("success", func(t *testing.T) {
		lc := fxtest.NewLifecycle(t)
		db, err := provideDB(lc, config.JudgeWorker{DatabaseURL: dsn})
		require.NoError(t, err)
		require.NotNil(t, db)
		lc.RequireStart()
		lc.RequireStop()
	})

	t.Run("connect error", func(t *testing.T) {
		lc := fxtest.NewLifecycle(t)
		_, err := provideDB(
			lc,
			config.JudgeWorker{
				DatabaseURL: "host=127.0.0.1 port=1 user=x password=x dbname=x sslmode=disable connect_timeout=1",
			},
		)
		require.Error(t, err)
	})
}

func TestProvideEvaluator(t *testing.T) {
	service := inmocks.NewMockEvaluationService(t)
	evaluator := provideEvaluator(config.JudgeWorker{TimeLimitSecs: 3}, service)
	require.NotNil(t, evaluator)
}

func TestProvideWorker(t *testing.T) {
	service := inmocks.NewMockEvaluationService(t)
	evaluator := provideEvaluator(config.JudgeWorker{TimeLimitSecs: 3}, service)
	worker := provideWorker(
		config.JudgeWorker{WorkerID: "test-worker", Concurrency: 2},
		outmocks.NewMockEvaluationJobQueue(t),
		evaluator,
	)
	require.NotNil(t, worker)
}

func TestRegisterLifecycle(t *testing.T) {
	lc := fxtest.NewLifecycle(t)
	service := evaluationServiceFunc(func(context.Context, uuid.UUID, uuid.UUID, int) error { return nil })
	queue := outmocks.NewMockEvaluationJobQueue(t)
	queue.EXPECT().ClaimEvaluationJob(mock.Anything, "test-worker").Return(nil, nil).Maybe()
	worker := provideWorker(
		config.JudgeWorker{WorkerID: "test-worker", Concurrency: 1, TimeLimitSecs: 1},
		queue,
		provideEvaluator(config.JudgeWorker{TimeLimitSecs: 1}, service),
	)
	registerLifecycle(lc, worker)
	lc.RequireStart()
	lc.RequireStop()
}

type evaluationServiceFunc func(ctx context.Context, submissionID, userID uuid.UUID, timeLimitSecs int) error

func (f evaluationServiceFunc) EvaluateSubmission(
	ctx context.Context,
	submissionID, userID uuid.UUID,
	timeLimitSecs int,
) error {
	return f(ctx, submissionID, userID, timeLimitSecs)
}
