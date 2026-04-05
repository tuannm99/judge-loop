package diagent

import (
	"net/http"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
	"github.com/tuannm99/judge-loop/internal/adapter/http/localagent"
	"github.com/tuannm99/judge-loop/internal/config"
	"go.uber.org/fx/fxtest"
)

func TestNew(t *testing.T) {
	require.NotNil(t, New(config.LocalAgent{}))
}

func TestProvideHelpers(t *testing.T) {
	cfg := config.LocalAgent{
		ServerURL:    "http://localhost:8080",
		Port:         7070,
		UserID:       uuid.NewString(),
		RegistryPath: "./registry",
	}

	handler := provideHandler(cfg)
	require.NotNil(t, handler)

	server := provideHTTP(cfg, handler)
	require.Equal(t, "127.0.0.1:7070", server.Addr)
	require.NotNil(t, server.Handler)
}

func TestRegisterLifecycle(t *testing.T) {
	lc := fxtest.NewLifecycle(t)
	server := &http.Server{
		Addr: "127.0.0.1:0",
		Handler: localagent.NewRouter(
			localagent.NewHandler(localagent.NewAPIClient("http://example.com", "u1"), "u1", "./registry"),
		),
	}
	registerLifecycle(lc, server)
	lc.RequireStart()
	lc.RequireStop()
}
