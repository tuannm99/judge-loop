package localagent

import (
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

func newTestServer(t *testing.T, fn func(w http.ResponseWriter, r *http.Request)) *httptest.Server {
	t.Helper()
	return httptest.NewServer(http.HandlerFunc(fn))
}

func writeLocalRegistry(t *testing.T, provider string) string {
	t.Helper()

	registryPath := t.TempDir()
	require.NoError(t, os.Mkdir(filepath.Join(registryPath, "providers"), 0o755))
	index := `{"version":"v1","updated_at":"2026-03-31T00:00:00Z","manifests":[` +
		`{"name":"main","path":"providers/main.json"}]}`
	require.NoError(t, os.WriteFile(filepath.Join(registryPath, "index.json"), []byte(index), 0o644))
	require.NoError(t, os.WriteFile(filepath.Join(registryPath, "providers", "main.json"), []byte(provider), 0o644))
	return registryPath
}

type roundTripFunc func(*http.Request) (*http.Response, error)

func (f roundTripFunc) RoundTrip(r *http.Request) (*http.Response, error) { return f(r) }
