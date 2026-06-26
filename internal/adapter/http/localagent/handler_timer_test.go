package localagent

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
	localtimer "github.com/tuannm99/judge-loop/internal/infrastructure/localtimer"
)

func TestLocalHandlerStartTimer(t *testing.T) {
	gin.SetMode(gin.TestMode)

	server := newTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		_ = json.NewEncoder(w).Encode(TimerResponse{ID: uuid.NewString(), Active: true})
	})
	defer server.Close()

	h := NewHandler(NewAPIClient(server.URL, "user-1"), "user-1", "")
	problemID := uuid.New()
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(
		http.MethodPost,
		"/local/timer/start",
		bytes.NewBufferString(`{"problem_id":"`+problemID.String()+`"}`),
	)
	c.Request.Header.Set("Content-Type", "application/json")
	h.StartTimer(c)
	require.Equal(t, http.StatusOK, w.Code)
}

func TestLocalHandlerStartTimerIgnoresInvalidBodyAndProblemID(t *testing.T) {
	gin.SetMode(gin.TestMode)

	h := NewHandler(NewAPIClient("http://127.0.0.1:1", "user-1"), "user-1", "")
	h.client.http.Timeout = 50 * time.Millisecond
	cases := []struct {
		name string
		body string
	}{
		{name: "malformed body", body: "{"},
		{name: "bad problem id", body: `{"problem_id":"bad"}`},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			h.timer = localtimer.New()
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			c.Request = httptest.NewRequest(http.MethodPost, "/local/timer/start", bytes.NewBufferString(tc.body))
			c.Request.Header.Set("Content-Type", "application/json")
			h.StartTimer(c)
			require.Equal(t, http.StatusOK, w.Code)
		})
	}
}

func TestLocalHandlerStopTimer(t *testing.T) {
	gin.SetMode(gin.TestMode)

	server := newTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})
	defer server.Close()

	h := NewHandler(NewAPIClient(server.URL, "user-1"), "user-1", "")
	h.timer.Start(nil)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodPost, "/local/timer/stop", nil)
	h.StopTimer(c)
	require.Equal(t, http.StatusOK, w.Code)
}

func TestLocalHandlerStopTimerWithoutActiveTimer(t *testing.T) {
	gin.SetMode(gin.TestMode)

	h := NewHandler(NewAPIClient("http://127.0.0.1:1", "user-1"), "user-1", "")
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodPost, "/local/timer/stop", nil)
	h.StopTimer(c)
	require.Equal(t, http.StatusOK, w.Code)
}

func TestLocalHandlerCurrentTimerUsesLocalTimer(t *testing.T) {
	gin.SetMode(gin.TestMode)

	h := NewHandler(NewAPIClient("http://127.0.0.1:1", "user-1"), "user-1", "")
	h.timer.Start(nil)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "/local/timer/current", nil)
	h.CurrentTimer(c)
	require.Equal(t, http.StatusOK, w.Code)
	require.Contains(t, w.Body.String(), `"active":true`)
}

func TestLocalHandlerCurrentTimerFallsBackToServer(t *testing.T) {
	gin.SetMode(gin.TestMode)

	server := newTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		_ = json.NewEncoder(w).Encode(TimerResponse{Active: true})
	})
	defer server.Close()

	h := NewHandler(NewAPIClient(server.URL, "user-1"), "user-1", "")
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "/local/timer/current", nil)
	h.CurrentTimer(c)
	require.Equal(t, http.StatusOK, w.Code)
}

func TestLocalHandlerCurrentTimerReturnsInactiveWhenServerFails(t *testing.T) {
	gin.SetMode(gin.TestMode)

	h := NewHandler(NewAPIClient("http://127.0.0.1:1", "user-1"), "user-1", "")
	h.client.http.Timeout = 50 * time.Millisecond
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "/local/timer/current", nil)
	h.CurrentTimer(c)
	require.Equal(t, http.StatusOK, w.Code)
	require.Contains(t, w.Body.String(), `"active":false`)
}
