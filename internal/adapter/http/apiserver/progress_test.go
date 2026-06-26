package apiserver

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	inport "github.com/tuannm99/judge-loop/internal/port/in"
	inmocks "github.com/tuannm99/judge-loop/internal/port/in/mocks"
	outport "github.com/tuannm99/judge-loop/internal/port/out"
)

func TestProgressAPIGetProgressToday(t *testing.T) {
	gin.SetMode(gin.TestMode)

	cases := []struct {
		name       string
		result     inport.ProgressToday
		err        error
		wantStatus int
	}{
		{
			name: "success",
			result: inport.ProgressToday{
				Date:             "2026-03-31",
				Solved:           2,
				Attempted:        3,
				TimeSpentMinutes: 4,
				Streak:           5,
			},
			wantStatus: http.StatusOK,
		},
		{
			name:       "service error",
			err:        errors.New("boom"),
			wantStatus: http.StatusInternalServerError,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			userID := uuid.New()
			service := inmocks.NewMockProgressService(t)
			api := New(nil, nil, service, nil, nil, nil, nil, userID)

			service.EXPECT().GetProgressToday(mock.Anything, userID).Return(tc.result, tc.err).Once()
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			c.Request = httptest.NewRequest(http.MethodGet, "/api/progress/today", nil)
			api.Progress.GetProgressToday(c)
			require.Equal(t, tc.wantStatus, w.Code)
		})
	}
}

func TestProgressAPIGetStreak(t *testing.T) {
	gin.SetMode(gin.TestMode)
	lastPracticed := time.Now()

	cases := []struct {
		name       string
		result     outport.StreakInfo
		err        error
		wantStatus int
	}{
		{
			name:       "success",
			result:     outport.StreakInfo{Current: 2, Longest: 4, LastPracticed: &lastPracticed},
			wantStatus: http.StatusOK,
		},
		{
			name:       "service error",
			err:        errors.New("boom"),
			wantStatus: http.StatusInternalServerError,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			userID := uuid.New()
			service := inmocks.NewMockProgressService(t)
			api := New(nil, nil, service, nil, nil, nil, nil, userID)

			service.EXPECT().GetStreak(mock.Anything, userID).Return(tc.result, tc.err).Once()
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			c.Request = httptest.NewRequest(http.MethodGet, "/api/streak", nil)
			api.Progress.GetStreak(c)
			require.Equal(t, tc.wantStatus, w.Code)
		})
	}
}

func TestProgressAPIGetGoalProgress(t *testing.T) {
	gin.SetMode(gin.TestMode)

	cases := []struct {
		name       string
		result     inport.GoalProgress
		err        error
		wantStatus int
	}{
		{
			name:       "success",
			result:     inport.GoalProgress{Solved: 12, Total: inport.GoalTotal, Remaining: inport.GoalTotal - 12},
			wantStatus: http.StatusOK,
		},
		{
			name:       "service error",
			err:        errors.New("boom"),
			wantStatus: http.StatusInternalServerError,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			userID := uuid.New()
			service := inmocks.NewMockProgressService(t)
			api := New(nil, nil, service, nil, nil, nil, nil, userID)

			service.EXPECT().GetGoalProgress(mock.Anything, userID).Return(tc.result, tc.err).Once()
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			c.Request = httptest.NewRequest(http.MethodGet, "/api/progress/goal", nil)
			api.Progress.GetGoalProgress(c)
			require.Equal(t, tc.wantStatus, w.Code)
		})
	}
}
