package localagent

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
	"github.com/tuannm99/judge-loop/internal/domain"
	localtimer "github.com/tuannm99/judge-loop/internal/infrastructure/localtimer"
)

func newTestServer(t *testing.T, fn func(w http.ResponseWriter, r *http.Request)) *httptest.Server {
	t.Helper()
	return httptest.NewServer(http.HandlerFunc(fn))
}

func TestAPIClient(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("successful requests", func(t *testing.T) {
		server := newTestServer(t, func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			switch r.URL.Path {
			case "/api/progress/today":
				_ = json.NewEncoder(w).Encode(ProgressResponse{Solved: 2, Streak: 3})
			case "/api/timers/start":
				_ = json.NewEncoder(w).Encode(TimerResponse{ID: uuid.NewString(), Active: true})
			case "/api/timers/stop":
				w.WriteHeader(http.StatusOK)
			case "/api/timers/current":
				_ = json.NewEncoder(w).Encode(TimerResponse{Active: true})
			case "/api/submissions":
				_ = json.NewEncoder(w).Encode(SubmitResponse{SubmissionID: "sub-1", Status: "pending"})
			case "/api/submissions/sub-1":
				_ = json.NewEncoder(w).Encode(SubmissionStatusResponse{ID: "sub-1", Status: "accepted"})
			case "/api/registry/sync":
				_ = json.NewEncoder(w).Encode(RegistrySyncResponse{Version: "v1", Synced: 3})
			default:
				http.NotFound(w, r)
			}
		})
		defer server.Close()

		client := NewAPIClient(server.URL, "user-1")
		progress, err := client.ProgressToday(context.Background())
		require.NoError(t, err)
		require.Equal(t, 2, progress.Solved)

		timerStart, err := client.StartTimer(context.Background(), "problem-1")
		require.NoError(t, err)
		require.NotEmpty(t, timerStart.ID)
		require.NoError(t, client.StopTimer(context.Background()))

		timer, err := client.CurrentTimer(context.Background())
		require.NoError(t, err)
		require.True(t, timer.Active)

		sub, err := client.Submit(context.Background(), SubmitRequest{ProblemID: "p1", Language: "go", Code: "code"})
		require.NoError(t, err)
		require.Equal(t, "sub-1", sub.SubmissionID)

		status, err := client.GetSubmission(context.Background(), "sub-1")
		require.NoError(t, err)
		require.Equal(t, "accepted", status.Status)

		syncResp, err := client.RegistrySync(context.Background(), RegistrySyncRequest{Version: "v1"})
		require.NoError(t, err)
		require.Equal(t, 3, syncResp.Synced)
	})

	t.Run("do handles marshal status decode and network errors", func(t *testing.T) {
		server := newTestServer(t, func(w http.ResponseWriter, r *http.Request) {
			switch r.URL.Path {
			case "/bad-status":
				http.Error(w, "bad", http.StatusBadGateway)
			case "/bad-json":
				_, _ = w.Write([]byte("{"))
			default:
				w.WriteHeader(http.StatusOK)
			}
		})
		defer server.Close()

		client := NewAPIClient(server.URL, "user-1")

		err := client.do(context.Background(), http.MethodPost, "/marshal", map[string]any{"bad": make(chan int)}, nil)
		require.Error(t, err)
		require.Contains(t, err.Error(), "marshal request")

		err = client.do(context.Background(), http.MethodGet, "/bad-status", nil, nil)
		require.Error(t, err)
		require.Contains(t, err.Error(), "502")

		var out map[string]any
		err = client.do(context.Background(), http.MethodGet, "/bad-json", nil, &out)
		require.Error(t, err)
		require.Contains(t, err.Error(), "decode response")

		client.baseURL = "://bad-url"
		err = client.do(context.Background(), http.MethodGet, "/x", nil, nil)
		require.Error(t, err)
		require.Contains(t, err.Error(), "build request")
	})

	t.Run("network error", func(t *testing.T) {
		client := NewAPIClient("http://127.0.0.1:1", "user-1")
		client.http.Timeout = 50 * time.Millisecond
		_, err := client.ProgressToday(context.Background())
		require.Error(t, err)
	})
}

func TestNewHandler(t *testing.T) {
	h := NewHandler(&APIClient{}, "user-1", "/tmp/registry")
	require.NotNil(t, h.client)
	require.NotNil(t, h.timer)
	require.Equal(t, "user-1", h.userID)
}

func TestLocalHandlerStatusAndSubmit(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("status uses server response and local fallback", func(t *testing.T) {
		server := newTestServer(t, func(w http.ResponseWriter, r *http.Request) {
			if r.URL.Path == "/api/progress/today" {
				_ = json.NewEncoder(w).Encode(ProgressResponse{Solved: 1, Streak: 2})
				return
			}
			http.Error(w, "missing", http.StatusNotFound)
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

		h.client.baseURL = "http://127.0.0.1:1"
		h.client.http.Timeout = 50 * time.Millisecond
		w = httptest.NewRecorder()
		c, _ = gin.CreateTestContext(w)
		c.Request = httptest.NewRequest(http.MethodGet, "/local/status/today", nil)
		h.StatusToday(c)
		require.Equal(t, http.StatusOK, w.Code)
		require.Contains(t, w.Body.String(), `"server_error"`)
	})

	t.Run("build status message cases", func(t *testing.T) {
		require.Equal(t, "Timer running — keep it up!", buildStatusMessage(1, true))
		require.Equal(t, "Good work today! Come back tomorrow.", buildStatusMessage(1, false))
		require.Equal(t, "Timer is running. Submit when ready.", buildStatusMessage(0, true))
		require.Equal(t, "No practice yet today. Start a session!", buildStatusMessage(0, false))
	})

	t.Run("submit handler", func(t *testing.T) {
		var submitted SubmitRequest
		server := newTestServer(t, func(w http.ResponseWriter, r *http.Request) {
			require.Equal(t, "/api/submissions", r.URL.Path)
			require.NoError(t, json.NewDecoder(r.Body).Decode(&submitted))
			_ = json.NewEncoder(w).Encode(SubmitResponse{SubmissionID: "sub-1", Status: "pending"})
		})
		defer server.Close()

		h := NewHandler(NewAPIClient(server.URL, "user-1"), "user-1", "")
		h.timer.Start(nil)

		body := `{"problem_id":"p1","language":"go","code":"fmt.Println(1)"}`
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest(http.MethodPost, "/local/submit", bytes.NewBufferString(body))
		c.Request.Header.Set("Content-Type", "application/json")
		h.Submit(c)
		require.Equal(t, http.StatusCreated, w.Code)
		require.Empty(t, submitted.SessionID)

		w = httptest.NewRecorder()
		c, _ = gin.CreateTestContext(w)
		c.Request = httptest.NewRequest(http.MethodPost, "/local/submit", bytes.NewBufferString("{"))
		c.Request.Header.Set("Content-Type", "application/json")
		h.Submit(c)
		require.Equal(t, http.StatusBadRequest, w.Code)

		h.client.baseURL = "http://127.0.0.1:1"
		h.client.http.Timeout = 50 * time.Millisecond
		w = httptest.NewRecorder()
		c, _ = gin.CreateTestContext(w)
		c.Request = httptest.NewRequest(http.MethodPost, "/local/submit", bytes.NewBufferString(body))
		c.Request.Header.Set("Content-Type", "application/json")
		h.Submit(c)
		require.Equal(t, http.StatusBadGateway, w.Code)
	})

	t.Run("submit attaches server timer id only", func(t *testing.T) {
		serverTimerID := uuid.New()
		submitted := make(chan SubmitRequest, 1)
		server := newTestServer(t, func(w http.ResponseWriter, r *http.Request) {
			switch r.URL.Path {
			case "/api/timers/start":
				_ = json.NewEncoder(w).Encode(TimerResponse{ID: serverTimerID.String(), Active: true})
			case "/api/submissions":
				var req SubmitRequest
				require.NoError(t, json.NewDecoder(r.Body).Decode(&req))
				submitted <- req
				_ = json.NewEncoder(w).Encode(SubmitResponse{SubmissionID: "sub-1", Status: "pending"})
			default:
				http.NotFound(w, r)
			}
		})
		defer server.Close()

		h := NewHandler(NewAPIClient(server.URL, "user-1"), "user-1", "")

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest(http.MethodPost, "/local/timer/start", bytes.NewBufferString(`{}`))
		c.Request.Header.Set("Content-Type", "application/json")
		h.StartTimer(c)
		require.Equal(t, http.StatusOK, w.Code)

		require.Eventually(t, func() bool {
			active := h.timer.Active()
			return active != nil && active.ServerID != nil && *active.ServerID == serverTimerID
		}, time.Second, 10*time.Millisecond)

		body := `{"problem_id":"p1","language":"go","code":"fmt.Println(1)"}`
		w = httptest.NewRecorder()
		c, _ = gin.CreateTestContext(w)
		c.Request = httptest.NewRequest(http.MethodPost, "/local/submit", bytes.NewBufferString(body))
		c.Request.Header.Set("Content-Type", "application/json")
		h.Submit(c)
		require.Equal(t, http.StatusCreated, w.Code)

		select {
		case req := <-submitted:
			require.Equal(t, serverTimerID.String(), req.SessionID)
		case <-time.After(time.Second):
			t.Fatal("submit request not captured")
		}
	})

	t.Run("get submission status", func(t *testing.T) {
		server := newTestServer(t, func(w http.ResponseWriter, r *http.Request) {
			_ = json.NewEncoder(w).Encode(SubmissionStatusResponse{ID: "sub-1", Status: "accepted"})
		})
		defer server.Close()

		h := NewHandler(NewAPIClient(server.URL, "user-1"), "user-1", "")
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Params = gin.Params{{Key: "id", Value: "sub-1"}}
		c.Request = httptest.NewRequest(http.MethodGet, "/local/submissions/sub-1", nil)
		h.GetSubmissionStatus(c)
		require.Equal(t, http.StatusOK, w.Code)

		h.client.baseURL = "http://127.0.0.1:1"
		h.client.http.Timeout = 50 * time.Millisecond
		w = httptest.NewRecorder()
		c, _ = gin.CreateTestContext(w)
		c.Params = gin.Params{{Key: "id", Value: "sub-1"}}
		c.Request = httptest.NewRequest(http.MethodGet, "/local/submissions/sub-1", nil)
		h.GetSubmissionStatus(c)
		require.Equal(t, http.StatusBadGateway, w.Code)
	})

	t.Run("problem proxies", func(t *testing.T) {
		server := newTestServer(t, func(w http.ResponseWriter, r *http.Request) {
			switch r.URL.Path {
			case "/api/problems":
				require.Equal(t, "20", r.URL.Query().Get("limit"))
				_ = json.NewEncoder(w).Encode(gin.H{"problems": []gin.H{}, "total": 0})
			case "/api/problems/suggest":
				_ = json.NewEncoder(w).Encode(gin.H{"id": "p1", "title": "Two Sum"})
			case "/api/problems/two-sum":
				_ = json.NewEncoder(w).Encode(gin.H{"id": "p1", "slug": "two-sum"})
			default:
				http.NotFound(w, r)
			}
		})
		defer server.Close()

		h := NewHandler(NewAPIClient(server.URL, "user-1"), "user-1", "")

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest(http.MethodGet, "/local/problems?limit=20", nil)
		h.ListProblems(c)
		require.Equal(t, http.StatusOK, w.Code)
		require.Contains(t, w.Body.String(), `"total":0`)

		w = httptest.NewRecorder()
		c, _ = gin.CreateTestContext(w)
		c.Request = httptest.NewRequest(http.MethodGet, "/local/problems/suggest", nil)
		h.SuggestProblem(c)
		require.Equal(t, http.StatusOK, w.Code)

		w = httptest.NewRecorder()
		c, _ = gin.CreateTestContext(w)
		c.Params = gin.Params{{Key: "id", Value: "two-sum"}}
		c.Request = httptest.NewRequest(http.MethodGet, "/local/problems/two-sum", nil)
		h.GetProblem(c)
		require.Equal(t, http.StatusOK, w.Code)
	})
}

func TestLocalHandlerTimerAndSync(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("timer lifecycle", func(t *testing.T) {
		server := newTestServer(t, func(w http.ResponseWriter, r *http.Request) {
			if r.URL.Path == "/api/timers/start" {
				_ = json.NewEncoder(w).Encode(TimerResponse{ID: uuid.NewString(), Active: true})
				return
			}
			w.WriteHeader(http.StatusOK)
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

		w = httptest.NewRecorder()
		c, _ = gin.CreateTestContext(w)
		c.Request = httptest.NewRequest(http.MethodGet, "/local/timer/current", nil)
		h.CurrentTimer(c)
		require.Equal(t, http.StatusOK, w.Code)
		require.Contains(t, w.Body.String(), `"active":true`)

		w = httptest.NewRecorder()
		c, _ = gin.CreateTestContext(w)
		c.Request = httptest.NewRequest(http.MethodPost, "/local/timer/stop", nil)
		h.StopTimer(c)
		require.Equal(t, http.StatusOK, w.Code)

		w = httptest.NewRecorder()
		c, _ = gin.CreateTestContext(w)
		c.Request = httptest.NewRequest(http.MethodPost, "/local/timer/stop", nil)
		h.StopTimer(c)
		require.Equal(t, http.StatusOK, w.Code)

		server.Config.Handler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			_ = json.NewEncoder(w).Encode(TimerResponse{Active: true})
		})
		w = httptest.NewRecorder()
		c, _ = gin.CreateTestContext(w)
		c.Request = httptest.NewRequest(http.MethodGet, "/local/timer/current", nil)
		h.CurrentTimer(c)
		require.Equal(t, http.StatusOK, w.Code)

		h.client.baseURL = "http://127.0.0.1:1"
		h.client.http.Timeout = 50 * time.Millisecond
		w = httptest.NewRecorder()
		c, _ = gin.CreateTestContext(w)
		c.Request = httptest.NewRequest(http.MethodGet, "/local/timer/current", nil)
		h.CurrentTimer(c)
		require.Equal(t, http.StatusOK, w.Code)
		require.Contains(t, w.Body.String(), `"active":false`)

		w = httptest.NewRecorder()
		c, _ = gin.CreateTestContext(w)
		c.Request = httptest.NewRequest(http.MethodPost, "/local/timer/start", bytes.NewBufferString("{"))
		c.Request.Header.Set("Content-Type", "application/json")
		h.StartTimer(c)
		require.Equal(t, http.StatusOK, w.Code)

		h.timer = localtimer.New()
		w = httptest.NewRecorder()
		c, _ = gin.CreateTestContext(w)
		c.Request = httptest.NewRequest(
			http.MethodPost,
			"/local/timer/start",
			bytes.NewBufferString(`{"problem_id":"bad"}`),
		)
		c.Request.Header.Set("Content-Type", "application/json")
		h.StartTimer(c)
		require.Equal(t, http.StatusOK, w.Code)
	})

	t.Run("sync handler", func(t *testing.T) {
		registryPath := t.TempDir()
		require.NoError(t, os.Mkdir(filepath.Join(registryPath, "providers"), 0o755))
		index := `{"version":"v1","updated_at":"2026-03-31T00:00:00Z","manifests":[{"name":"main","path":"providers/main.json"}]}`
		provider := `{"problems":[{"slug":"two-sum","title":"Two Sum","difficulty":"easy","provider":"leetcode","external_id":"1"}]}`
		require.NoError(t, os.WriteFile(filepath.Join(registryPath, "index.json"), []byte(index), 0o644))
		require.NoError(t, os.WriteFile(filepath.Join(registryPath, "providers", "main.json"), []byte(provider), 0o644))

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

		h.registryPath = ""
		w = httptest.NewRecorder()
		c, _ = gin.CreateTestContext(w)
		c.Request = httptest.NewRequest(http.MethodPost, "/local/sync", nil)
		h.Sync(c)
		require.Equal(t, http.StatusServiceUnavailable, w.Code)

		h.registryPath = t.TempDir()
		w = httptest.NewRecorder()
		c, _ = gin.CreateTestContext(w)
		c.Request = httptest.NewRequest(http.MethodPost, "/local/sync", nil)
		h.Sync(c)
		require.Equal(t, http.StatusInternalServerError, w.Code)

		badRegistry := t.TempDir()
		require.NoError(t, os.Mkdir(filepath.Join(badRegistry, "providers"), 0o755))
		require.NoError(t, os.WriteFile(filepath.Join(badRegistry, "index.json"), []byte(index), 0o644))
		require.NoError(t, os.WriteFile(filepath.Join(badRegistry, "providers", "main.json"), []byte("{"), 0o644))
		h.registryPath = badRegistry
		w = httptest.NewRecorder()
		c, _ = gin.CreateTestContext(w)
		c.Request = httptest.NewRequest(http.MethodPost, "/local/sync", nil)
		h.Sync(c)
		require.Equal(t, http.StatusInternalServerError, w.Code)

		h.registryPath = registryPath
		h.client.baseURL = "http://127.0.0.1:1"
		h.client.http.Timeout = 50 * time.Millisecond
		w = httptest.NewRecorder()
		c, _ = gin.CreateTestContext(w)
		c.Request = httptest.NewRequest(http.MethodPost, "/local/sync", nil)
		h.Sync(c)
		require.Equal(t, http.StatusBadGateway, w.Code)
	})
}

func TestDoWithCustomTransportError(t *testing.T) {
	client := NewAPIClient("http://example.com", "user-1")
	client.http = &http.Client{Transport: roundTripFunc(func(*http.Request) (*http.Response, error) {
		return nil, errors.New("transport")
	})}
	err := client.do(context.Background(), http.MethodGet, "/x", nil, nil)
	require.Error(t, err)
}

type roundTripFunc func(*http.Request) (*http.Response, error)

func (f roundTripFunc) RoundTrip(r *http.Request) (*http.Response, error) { return f(r) }
