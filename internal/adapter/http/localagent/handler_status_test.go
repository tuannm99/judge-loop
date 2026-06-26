package localagent

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLocalHandlerStatusToday(t *testing.T) {
	gin.SetMode(gin.TestMode)

	server := newTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/api/progress/today", r.URL.Path)
		_ = json.NewEncoder(w).Encode(ProgressResponse{Solved: 1, Streak: 2})
	})
	defer server.Close()

	h := NewHandler(NewAPIClient(server.URL, "user-1"), "user-1", "")
	h.timer.Start(nil)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "/local/status/today", nil)
	h.StatusToday(c)
	require.Equal(t, http.StatusOK, w.Code)
	require.Contains(t, w.Body.String(), `"practiced":true`)
}

func TestLocalHandlerStatusTodayFallsBackOnServerError(t *testing.T) {
	gin.SetMode(gin.TestMode)

	h := NewHandler(NewAPIClient("http://127.0.0.1:1", "user-1"), "user-1", "")
	h.client.http.Timeout = 50 * time.Millisecond
	h.timer.Start(nil)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "/local/status/today", nil)
	h.StatusToday(c)
	require.Equal(t, http.StatusOK, w.Code)
	require.Contains(t, w.Body.String(), `"server_error"`)
}

func TestBuildStatusMessage(t *testing.T) {
	cases := []struct {
		name   string
		solved int
		active bool
		want   string
	}{
		{name: "solved active", solved: 1, active: true, want: "Timer running — keep it up!"},
		{name: "solved inactive", solved: 1, want: "Good work today! Come back tomorrow."},
		{name: "unsolved active", active: true, want: "Timer is running. Submit when ready."},
		{name: "unsolved inactive", want: "No practice yet today. Start a session!"},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			require.Equal(t, tc.want, buildStatusMessage(tc.solved, tc.active))
		})
	}
}
