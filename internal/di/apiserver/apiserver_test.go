package diserver

import (
	"context"
	"fmt"
	"net/http"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
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

func TestNew(t *testing.T) {
	require.NotNil(t, New(config.APIServer{}))
}

func TestProvideDB(t *testing.T) {
	dsn := newPostgresDSN(t)

	t.Run("success", func(t *testing.T) {
		lc := fxtest.NewLifecycle(t)
		db, err := provideDB(lc, config.APIServer{DatabaseURL: dsn})
		require.NoError(t, err)
		require.NotNil(t, db)
		lc.RequireStart()
		lc.RequireStop()
	})

	t.Run("connect error", func(t *testing.T) {
		lc := fxtest.NewLifecycle(t)
		_, err := provideDB(
			lc,
			config.APIServer{
				DatabaseURL: "host=127.0.0.1 port=1 user=x password=x dbname=x sslmode=disable connect_timeout=1",
			},
		)
		require.Error(t, err)
	})
}

func TestProvideHTTPAndQueue(t *testing.T) {
	cfg := config.APIServer{Port: "8080", UserID: uuid.NewString()}

	server, err := provideHTTP(
		cfg,
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

	_, err = provideHTTP(
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
	client := provideQueueClient(lc, config.APIServer{RedisURL: "redis://localhost:6379"})
	require.NotNil(t, client)
	lc.RequireStart()
	lc.RequireStop()
}

func TestRegisterLifecycle(t *testing.T) {
	lc := fxtest.NewLifecycle(t)
	server := &http.Server{Addr: "127.0.0.1:0", Handler: http.NewServeMux()}
	registerLifecycle(lc, server)
	lc.RequireStart()
	lc.RequireStop()
}
