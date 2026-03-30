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

func TestProgressHandlers(t *testing.T) {
	gin.SetMode(gin.TestMode)
	userID := uuid.New()
	service := inmocks.NewMockAPIService(t)
	api := NewWithService(service, userID)

	service.EXPECT().
		GetProgressToday(mock.Anything, userID).
		Return(inport.ProgressToday{Date: "2026-03-31", Solved: 2, Attempted: 3, TimeSpentMinutes: 4, Streak: 5}, nil).
		Once()
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "/api/progress/today", nil)
	api.Progress.GetProgressToday(c)
	require.Equal(t, http.StatusOK, w.Code)

	service.EXPECT().GetProgressToday(mock.Anything, userID).Return(inport.ProgressToday{}, errors.New("boom")).Once()
	w = httptest.NewRecorder()
	c, _ = gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "/api/progress/today", nil)
	api.Progress.GetProgressToday(c)
	require.Equal(t, http.StatusInternalServerError, w.Code)

	lastPracticed := time.Now()
	service.EXPECT().
		GetStreak(mock.Anything, userID).
		Return(outport.StreakInfo{Current: 2, Longest: 4, LastPracticed: &lastPracticed}, nil).
		Once()
	w = httptest.NewRecorder()
	c, _ = gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "/api/streak", nil)
	api.Progress.GetStreak(c)
	require.Equal(t, http.StatusOK, w.Code)

	service.EXPECT().GetStreak(mock.Anything, userID).Return(outport.StreakInfo{}, errors.New("boom")).Once()
	w = httptest.NewRecorder()
	c, _ = gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "/api/streak", nil)
	api.Progress.GetStreak(c)
	require.Equal(t, http.StatusInternalServerError, w.Code)
}
