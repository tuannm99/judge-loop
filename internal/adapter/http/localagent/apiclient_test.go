package localagent

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/url"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAPIClientProgressToday(t *testing.T) {
	server := newTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/api/progress/today", r.URL.Path)
		_ = json.NewEncoder(w).Encode(ProgressResponse{Solved: 2, Streak: 3})
	})
	defer server.Close()

	progress, err := NewAPIClient(server.URL, "user-1").ProgressToday(context.Background())
	require.NoError(t, err)
	require.Equal(t, 2, progress.Solved)
}

func TestAPIClientStartTimer(t *testing.T) {
	server := newTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/api/timers/start", r.URL.Path)
		_ = json.NewEncoder(w).Encode(TimerResponse{ID: uuid.NewString(), Active: true})
	})
	defer server.Close()

	timer, err := NewAPIClient(server.URL, "user-1").StartTimer(context.Background(), "problem-1")
	require.NoError(t, err)
	require.NotEmpty(t, timer.ID)
}

func TestAPIClientStopTimer(t *testing.T) {
	server := newTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/api/timers/stop", r.URL.Path)
		w.WriteHeader(http.StatusOK)
	})
	defer server.Close()

	require.NoError(t, NewAPIClient(server.URL, "user-1").StopTimer(context.Background()))
}

func TestAPIClientCurrentTimer(t *testing.T) {
	server := newTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/api/timers/current", r.URL.Path)
		_ = json.NewEncoder(w).Encode(TimerResponse{Active: true})
	})
	defer server.Close()

	timer, err := NewAPIClient(server.URL, "user-1").CurrentTimer(context.Background())
	require.NoError(t, err)
	require.True(t, timer.Active)
}

func TestAPIClientSubmit(t *testing.T) {
	server := newTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/api/submissions", r.URL.Path)
		_ = json.NewEncoder(w).Encode(SubmitResponse{SubmissionID: "sub-1", Status: "pending"})
	})
	defer server.Close()

	sub, err := NewAPIClient(server.URL, "user-1").
		Submit(context.Background(), SubmitRequest{ProblemID: "p1", Language: "go", Code: "code"})
	require.NoError(t, err)
	require.Equal(t, "sub-1", sub.SubmissionID)
}

func TestAPIClientGetSubmission(t *testing.T) {
	server := newTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/api/submissions/sub-1", r.URL.Path)
		_ = json.NewEncoder(w).Encode(SubmissionStatusResponse{ID: "sub-1", Status: "accepted"})
	})
	defer server.Close()

	status, err := NewAPIClient(server.URL, "user-1").GetSubmission(context.Background(), "sub-1")
	require.NoError(t, err)
	require.Equal(t, "accepted", status.Status)
}

func TestAPIClientProxyGet(t *testing.T) {
	server := newTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/api/problems", r.URL.Path)
		assert.Equal(t, "20", r.URL.Query().Get("limit"))
		_ = json.NewEncoder(w).Encode(map[string]any{"total": 0})
	})
	defer server.Close()

	raw, err := NewAPIClient(server.URL, "user-1").
		ProxyGet(context.Background(), "/api/problems", url.Values{"limit": []string{"20"}})
	require.NoError(t, err)
	require.Contains(t, string(raw), `"total":0`)
}

func TestAPIClientRegistrySync(t *testing.T) {
	server := newTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/api/registry/sync", r.URL.Path)
		_ = json.NewEncoder(w).Encode(RegistrySyncResponse{Version: "v1", Synced: 3})
	})
	defer server.Close()

	syncResp, err := NewAPIClient(server.URL, "user-1").
		RegistrySync(context.Background(), RegistrySyncRequest{Version: "v1"})
	require.NoError(t, err)
	require.Equal(t, 3, syncResp.Synced)
}

func TestAPIClientDoErrors(t *testing.T) {
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
	cases := []struct {
		name     string
		run      func() error
		contains string
	}{
		{
			name: "marshal error",
			run: func() error {
				return client.do(
					context.Background(),
					http.MethodPost,
					"/marshal",
					map[string]any{"bad": make(chan int)},
					nil,
				)
			},
			contains: "marshal request",
		},
		{
			name:     "bad status",
			run:      func() error { return client.do(context.Background(), http.MethodGet, "/bad-status", nil, nil) },
			contains: "502",
		},
		{
			name: "decode error",
			run: func() error {
				var out map[string]any
				return client.do(context.Background(), http.MethodGet, "/bad-json", nil, &out)
			},
			contains: "decode response",
		},
		{
			name: "build request error",
			run: func() error {
				client.baseURL = "://bad-url"
				return client.do(context.Background(), http.MethodGet, "/x", nil, nil)
			},
			contains: "build request",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			err := tc.run()
			require.Error(t, err)
			require.Contains(t, err.Error(), tc.contains)
		})
	}
}

func TestAPIClientProgressTodayReturnsNetworkError(t *testing.T) {
	client := NewAPIClient("http://127.0.0.1:1", "user-1")
	client.http.Timeout = 50 * time.Millisecond
	_, err := client.ProgressToday(context.Background())
	require.Error(t, err)
}

func TestDoWithCustomTransportError(t *testing.T) {
	client := NewAPIClient("http://example.com", "user-1")
	client.http = &http.Client{Transport: roundTripFunc(func(*http.Request) (*http.Response, error) {
		return nil, errors.New("transport")
	})}
	err := client.do(context.Background(), http.MethodGet, "/x", nil, nil)
	require.Error(t, err)
}
