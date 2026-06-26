package apiserver

import (
	"bytes"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"github.com/tuannm99/judge-loop/internal/domain"
	inmocks "github.com/tuannm99/judge-loop/internal/port/in/mocks"
)

func TestTimersAPIStartTimer(t *testing.T) {
	gin.SetMode(gin.TestMode)

	problemID := uuid.New()
	timerID := uuid.New()
	startedAt := time.Now().Add(-2 * time.Minute)
	cases := []struct {
		name       string
		body       string
		problemID  *uuid.UUID
		timer      *domain.TimerSession
		err        error
		wantStatus int
	}{
		{
			name:       "success",
			body:       `{"problem_id":"` + problemID.String() + `"}`,
			problemID:  &problemID,
			timer:      &domain.TimerSession{ID: timerID, StartedAt: startedAt, ProblemID: &problemID},
			wantStatus: http.StatusOK,
		},
		{
			name:       "invalid problem id falls back to nil and returns service error",
			body:       `{"problem_id":"bad"}`,
			err:        errors.New("boom"),
			wantStatus: http.StatusInternalServerError,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			userID := uuid.New()
			service := inmocks.NewMockTimerService(t)
			api := New(nil, nil, nil, service, nil, nil, nil, userID)

			service.EXPECT().StartTimer(mock.Anything, userID, tc.problemID).Return(tc.timer, tc.err).Once()
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			c.Request = httptest.NewRequest(http.MethodPost, "/api/timers/start", bytes.NewBufferString(tc.body))
			c.Request.Header.Set("Content-Type", "application/json")
			api.Timers.StartTimer(c)
			require.Equal(t, tc.wantStatus, w.Code)
		})
	}
}

func TestTimersAPIStopTimer(t *testing.T) {
	gin.SetMode(gin.TestMode)

	cases := []struct {
		name       string
		timer      *domain.TimerSession
		err        error
		wantStatus int
	}{
		{name: "no active timer", wantStatus: http.StatusOK},
		{name: "success", timer: &domain.TimerSession{ElapsedSecs: 12}, wantStatus: http.StatusOK},
		{name: "service error", err: errors.New("boom"), wantStatus: http.StatusInternalServerError},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			userID := uuid.New()
			service := inmocks.NewMockTimerService(t)
			api := New(nil, nil, nil, service, nil, nil, nil, userID)

			service.EXPECT().StopTimer(mock.Anything, userID).Return(tc.timer, tc.err).Once()
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			c.Request = httptest.NewRequest(http.MethodPost, "/api/timers/stop", nil)
			api.Timers.StopTimer(c)
			require.Equal(t, tc.wantStatus, w.Code)
		})
	}
}

func TestTimersAPICurrentTimer(t *testing.T) {
	gin.SetMode(gin.TestMode)

	problemID := uuid.New()
	timerID := uuid.New()
	startedAt := time.Now().Add(-2 * time.Minute)
	cases := []struct {
		name       string
		timer      *domain.TimerSession
		err        error
		wantStatus int
	}{
		{name: "no active timer", wantStatus: http.StatusOK},
		{
			name:       "success",
			timer:      &domain.TimerSession{ID: timerID, StartedAt: startedAt, ProblemID: &problemID},
			wantStatus: http.StatusOK,
		},
		{name: "service error", err: errors.New("boom"), wantStatus: http.StatusInternalServerError},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			userID := uuid.New()
			service := inmocks.NewMockTimerService(t)
			api := New(nil, nil, nil, service, nil, nil, nil, userID)

			service.EXPECT().CurrentTimer(mock.Anything, userID).Return(tc.timer, tc.err).Once()
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			c.Request = httptest.NewRequest(http.MethodGet, "/api/timers/current", nil)
			api.Timers.CurrentTimer(c)
			require.Equal(t, tc.wantStatus, w.Code)
		})
	}
}
