package di

import (
	"context"
	"fmt"
	"net/http"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/hibiken/asynq"
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
	"github.com/tuannm99/judge-loop/internal/adapter/http/localagent"
	"github.com/tuannm99/judge-loop/internal/config"
	inmocks "github.com/tuannm99/judge-loop/internal/port/in/mocks"
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

func TestNewApps(t *testing.T) {
	require.NotNil(t, NewAPIServer(config.APIServer{}))
	require.NotNil(t, NewJudgeWorker(config.JudgeWorker{}))
	require.NotNil(t, NewLocalAgent(config.LocalAgent{}))
}

func TestProvideDBs(t *testing.T) {
	dsn := newPostgresDSN(t)

	t.Run("api server db success", func(t *testing.T) {
		lc := fxtest.NewLifecycle(t)
		db, err := provideAPIServerDB(lc, config.APIServer{DatabaseURL: dsn})
		require.NoError(t, err)
		require.NotNil(t, db)
		lc.RequireStart()
		lc.RequireStop()
	})

	t.Run("judge worker db success", func(t *testing.T) {
		lc := fxtest.NewLifecycle(t)
		db, err := provideJudgeWorkerDB(lc, config.JudgeWorker{DatabaseURL: dsn})
		require.NoError(t, err)
		require.NotNil(t, db)
		lc.RequireStart()
		lc.RequireStop()
	})

	t.Run("db connect error", func(t *testing.T) {
		lc := fxtest.NewLifecycle(t)
		_, err := provideAPIServerDB(lc, config.APIServer{DatabaseURL: "host=127.0.0.1 port=1 user=x password=x dbname=x sslmode=disable connect_timeout=1"})
		require.Error(t, err)
		_, err = provideJudgeWorkerDB(lc, config.JudgeWorker{DatabaseURL: "host=127.0.0.1 port=1 user=x password=x dbname=x sslmode=disable connect_timeout=1"})
		require.Error(t, err)
	})
}

func TestProvideAPIAndQueue(t *testing.T) {
	apiServerCfg := config.APIServer{Port: "8080", UserID: uuid.NewString()}

	server, err := provideAPI(
		apiServerCfg,
		inmocks.NewMockProblemService(t),
		inmocks.NewMockSubmissionService(t),
		inmocks.NewMockProgressService(t),
		inmocks.NewMockTimerService(t),
		inmocks.NewMockReviewService(t),
		inmocks.NewMockRegistryService(t),
	)
	require.NoError(t, err)
	require.Equal(t, ":8080", server.Addr)
	require.NotNil(t, server.Handler)

	_, err = provideAPI(
		config.APIServer{Port: "8080", UserID: "not-a-uuid"},
		inmocks.NewMockProblemService(t),
		inmocks.NewMockSubmissionService(t),
		inmocks.NewMockProgressService(t),
		inmocks.NewMockTimerService(t),
		inmocks.NewMockReviewService(t),
		inmocks.NewMockRegistryService(t),
	)
	require.Error(t, err)

	lc := fxtest.NewLifecycle(t)
	client := provideAPIQueueClient(lc, config.APIServer{RedisURL: "redis://localhost:6379"})
	require.NotNil(t, client)
	lc.RequireStart()
	lc.RequireStop()
}

func TestProvideWorkerHelpers(t *testing.T) {
	service := inmocks.NewMockEvaluationService(t)
	evaluator := provideEvaluator(config.JudgeWorker{TimeLimitSecs: 3}, service)
	require.NotNil(t, evaluator)

	server := provideWorkerServer(config.JudgeWorker{RedisURL: "redis://localhost:6379", Concurrency: 2})
	require.NotNil(t, server)

	mux := provideWorkerMux(evaluator)
	require.NotNil(t, mux)
}

func TestProvideLocalAgentHelpers(t *testing.T) {
	cfg := config.LocalAgent{
		ServerURL:    "http://localhost:8080",
		Port:         7070,
		UserID:       uuid.NewString(),
		RegistryPath: "./registry",
	}

	handler := provideLocalAgentHandler(cfg)
	require.NotNil(t, handler)

	server := provideLocalAgentHTTP(cfg, handler)
	require.Equal(t, "127.0.0.1:7070", server.Addr)
	require.NotNil(t, server.Handler)
}

func TestRegisterHTTPServerLifecycles(t *testing.T) {
	t.Run("api server lifecycle", func(t *testing.T) {
		lc := fxtest.NewLifecycle(t)
		server := &http.Server{Addr: "127.0.0.1:0", Handler: http.NewServeMux()}
		registerAPIServerLifecycle(lc, server)
		lc.RequireStart()
		lc.RequireStop()
	})

	t.Run("local agent lifecycle", func(t *testing.T) {
		lc := fxtest.NewLifecycle(t)
		server := &http.Server{Addr: "127.0.0.1:0", Handler: localagent.NewRouter(localagent.NewHandler(localagent.NewAPIClient("http://example.com", "u1"), "u1", "./registry"))}
		registerLocalAgentLifecycle(lc, server)
		lc.RequireStart()
		lc.RequireStop()
	})
}

func TestRegisterWorkerLifecycle(t *testing.T) {
	lc := fxtest.NewLifecycle(t)
	server := asynq.NewServer(asynq.RedisClientOpt{Addr: "localhost:0"}, asynq.Config{Concurrency: 1})
	mux := asynq.NewServeMux()
	registerWorkerLifecycle(lc, server, mux)
	lc.RequireStart()
	lc.RequireStop()
}
