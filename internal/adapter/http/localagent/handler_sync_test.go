package localagent

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
	"github.com/tuannm99/judge-loop/internal/domain"
)

func TestLocalHandlerSync(t *testing.T) {
	gin.SetMode(gin.TestMode)

	registryPath := writeLocalRegistry(t, `{"problems":[`+
		`{"slug":"two-sum","title":"Two Sum","difficulty":"easy","provider":"leetcode","external_id":"1"}]}`)
	server := newTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		_ = json.NewEncoder(w).Encode(RegistrySyncResponse{Version: "v1", Synced: 1})
	})
	defer server.Close()

	h := NewHandler(NewAPIClient(server.URL, "user-1"), "user-1", registryPath)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodPost, "/local/sync", nil)
	h.Sync(c)
	require.Equal(t, http.StatusOK, w.Code)
	require.Equal(
		t,
		[]domain.ManifestRef{{Name: "main", Path: "providers/main.json"}},
		toRefDTOs([]domain.ManifestRef{{Name: "main", Path: "providers/main.json"}}),
	)
}

func TestLocalHandlerSyncErrors(t *testing.T) {
	gin.SetMode(gin.TestMode)

	cases := []struct {
		name         string
		registryPath string
		baseURL      string
		wantStatus   int
	}{
		{name: "missing registry path", baseURL: "http://127.0.0.1:1", wantStatus: http.StatusServiceUnavailable},
		{
			name:         "missing index",
			registryPath: t.TempDir(),
			baseURL:      "http://127.0.0.1:1",
			wantStatus:   http.StatusInternalServerError,
		},
		{
			name:         "bad provider manifest",
			registryPath: writeLocalRegistry(t, "{"),
			baseURL:      "http://127.0.0.1:1",
			wantStatus:   http.StatusInternalServerError,
		},
		{
			name: "api server error",
			registryPath: writeLocalRegistry(
				t,
				`{"problems":[{"slug":"two-sum","title":"Two Sum","difficulty":"easy","provider":"leetcode","external_id":"1"}]}`,
			),
			baseURL:    "http://127.0.0.1:1",
			wantStatus: http.StatusBadGateway,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			h := NewHandler(NewAPIClient(tc.baseURL, "user-1"), "user-1", tc.registryPath)
			h.client.http.Timeout = 50 * time.Millisecond
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			c.Request = httptest.NewRequest(http.MethodPost, "/local/sync", nil)
			h.Sync(c)
			require.Equal(t, tc.wantStatus, w.Code)
		})
	}
}
