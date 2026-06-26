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
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLocalHandlerSubmit(t *testing.T) {
	gin.SetMode(gin.TestMode)

	var submitted SubmitRequest
	server := newTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/api/submissions", r.URL.Path)
		assert.NoError(t, json.NewDecoder(r.Body).Decode(&submitted))
		_ = json.NewEncoder(w).Encode(SubmitResponse{SubmissionID: "sub-1", Status: "pending"})
	})
	defer server.Close()

	h := NewHandler(NewAPIClient(server.URL, "user-1"), "user-1", "")
	h.timer.Start(nil)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(
		http.MethodPost,
		"/local/submit",
		bytes.NewBufferString(`{"problem_id":"p1","language":"go","code":"fmt.Println(1)"}`),
	)
	c.Request.Header.Set("Content-Type", "application/json")
	h.Submit(c)
	require.Equal(t, http.StatusCreated, w.Code)
	require.Empty(t, submitted.SessionID)
}

func TestLocalHandlerSubmitErrors(t *testing.T) {
	gin.SetMode(gin.TestMode)

	cases := []struct {
		name       string
		body       string
		baseURL    string
		wantStatus int
	}{
		{name: "bad request", body: "{", baseURL: "http://127.0.0.1:1", wantStatus: http.StatusBadRequest},
		{
			name:       "server error",
			body:       `{"problem_id":"p1","language":"go","code":"fmt.Println(1)"}`,
			baseURL:    "http://127.0.0.1:1",
			wantStatus: http.StatusBadGateway,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			h := NewHandler(NewAPIClient(tc.baseURL, "user-1"), "user-1", "")
			h.client.http.Timeout = 50 * time.Millisecond
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			c.Request = httptest.NewRequest(http.MethodPost, "/local/submit", bytes.NewBufferString(tc.body))
			c.Request.Header.Set("Content-Type", "application/json")
			h.Submit(c)
			require.Equal(t, tc.wantStatus, w.Code)
		})
	}
}

func TestLocalHandlerSubmitAttachesServerTimerID(t *testing.T) {
	gin.SetMode(gin.TestMode)

	serverTimerID := uuid.New()
	submitted := make(chan SubmitRequest, 1)
	server := newTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/api/timers/start":
			_ = json.NewEncoder(w).Encode(TimerResponse{ID: serverTimerID.String(), Active: true})
		case "/api/submissions":
			var req SubmitRequest
			assert.NoError(t, json.NewDecoder(r.Body).Decode(&req))
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

	w = httptest.NewRecorder()
	c, _ = gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(
		http.MethodPost,
		"/local/submit",
		bytes.NewBufferString(`{"problem_id":"p1","language":"go","code":"fmt.Println(1)"}`),
	)
	c.Request.Header.Set("Content-Type", "application/json")
	h.Submit(c)
	require.Equal(t, http.StatusCreated, w.Code)

	select {
	case req := <-submitted:
		require.Equal(t, serverTimerID.String(), req.SessionID)
	case <-time.After(time.Second):
		t.Fatal("submit request not captured")
	}
}
