package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

// APIClient wraps calls to the remote api-server.
// All methods return an error if the server is unreachable or returns a non-2xx status.
type APIClient struct {
	baseURL string
	userID  string
	http    *http.Client
}

// NewAPIClient creates a client pointing at the given api-server base URL.
func NewAPIClient(serverURL, userID string) *APIClient {
	return &APIClient{
		baseURL: serverURL,
		userID:  userID,
		http:    &http.Client{},
	}
}

// ProgressResponse mirrors GET /api/progress/today.
type ProgressResponse struct {
	Date             string `json:"date"`
	Solved           int    `json:"solved"`
	Attempted        int    `json:"attempted"`
	TimeSpentMinutes int    `json:"time_spent_minutes"`
	Streak           int    `json:"streak"`
}

// TimerResponse mirrors GET /api/timers/current.
type TimerResponse struct {
	Active         bool       `json:"active"`
	StartedAt      *time.Time `json:"started_at,omitempty"`
	ElapsedSeconds int        `json:"elapsed_seconds,omitempty"`
	ProblemID      *string    `json:"problem_id,omitempty"`
}

// SubmitRequest mirrors POST /api/submissions body.
type SubmitRequest struct {
	ProblemID string `json:"problem_id"`
	Language  string `json:"language"`
	Code      string `json:"code"`
	SessionID string `json:"session_id,omitempty"`
}

// SubmitResponse mirrors POST /api/submissions response.
type SubmitResponse struct {
	SubmissionID string `json:"submission_id"`
	Status       string `json:"status"`
}

// ProgressToday fetches today's practice summary from the api-server.
func (c *APIClient) ProgressToday(ctx context.Context) (*ProgressResponse, error) {
	var resp ProgressResponse
	if err := c.do(ctx, http.MethodGet, "/api/progress/today", nil, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

// StartTimer notifies the api-server that a timer has started.
func (c *APIClient) StartTimer(ctx context.Context, problemID string) error {
	body := map[string]string{"problem_id": problemID}
	return c.do(ctx, http.MethodPost, "/api/timers/start", body, nil)
}

// StopTimer notifies the api-server that the active timer has stopped.
func (c *APIClient) StopTimer(ctx context.Context) error {
	return c.do(ctx, http.MethodPost, "/api/timers/stop", map[string]any{}, nil)
}

// CurrentTimer fetches the active timer state from the api-server.
func (c *APIClient) CurrentTimer(ctx context.Context) (*TimerResponse, error) {
	var resp TimerResponse
	if err := c.do(ctx, http.MethodGet, "/api/timers/current", nil, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

// Submit forwards a submission to the api-server.
func (c *APIClient) Submit(ctx context.Context, req SubmitRequest) (*SubmitResponse, error) {
	var resp SubmitResponse
	if err := c.do(ctx, http.MethodPost, "/api/submissions", req, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

// do is the shared HTTP helper: marshals body, sends the request,
// and optionally decodes the JSON response into out.
func (c *APIClient) do(ctx context.Context, method, path string, body, out any) error {
	var bodyReader io.Reader
	if body != nil {
		b, err := json.Marshal(body)
		if err != nil {
			return fmt.Errorf("marshal request: %w", err)
		}
		bodyReader = bytes.NewReader(b)
	}

	req, err := http.NewRequestWithContext(ctx, method, c.baseURL+path, bodyReader)
	if err != nil {
		return fmt.Errorf("build request: %w", err)
	}
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	resp, err := c.http.Do(req)
	if err != nil {
		return err // network error: unreachable, timeout, etc.
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		return fmt.Errorf("api-server %s %s: %s", method, path, resp.Status)
	}

	if out != nil {
		if err := json.NewDecoder(resp.Body).Decode(out); err != nil {
			return fmt.Errorf("decode response: %w", err)
		}
	}
	return nil
}
