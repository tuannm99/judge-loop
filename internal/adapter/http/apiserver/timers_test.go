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

func TestTimerHandlers(t *testing.T) {
	gin.SetMode(gin.TestMode)
	userID := uuid.New()
	service := inmocks.NewMockTimerService(t)
	api := New(nil, nil, nil, service, nil, nil, userID)
	problemID := uuid.New()
	timerID := uuid.New()
	startedAt := time.Now().Add(-2 * time.Minute)

	service.EXPECT().
		StartTimer(mock.Anything, userID, &problemID).
		Return(&domain.TimerSession{ID: timerID, StartedAt: startedAt, ProblemID: &problemID}, nil).
		Once()
	body := `{"problem_id":"` + problemID.String() + `"}`
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodPost, "/api/timers/start", bytes.NewBufferString(body))
	c.Request.Header.Set("Content-Type", "application/json")
	api.Timers.StartTimer(c)
	require.Equal(t, http.StatusOK, w.Code)

	service.EXPECT().StartTimer(mock.Anything, userID, (*uuid.UUID)(nil)).Return(nil, errors.New("boom")).Once()
	w = httptest.NewRecorder()
	c, _ = gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodPost, "/api/timers/start", bytes.NewBufferString(`{"problem_id":"bad"}`))
	c.Request.Header.Set("Content-Type", "application/json")
	api.Timers.StartTimer(c)
	require.Equal(t, http.StatusInternalServerError, w.Code)

	service.EXPECT().StopTimer(mock.Anything, userID).Return(nil, nil).Once()
	w = httptest.NewRecorder()
	c, _ = gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodPost, "/api/timers/stop", nil)
	api.Timers.StopTimer(c)
	require.Equal(t, http.StatusOK, w.Code)

	service.EXPECT().StopTimer(mock.Anything, userID).Return(&domain.TimerSession{ElapsedSecs: 12}, nil).Once()
	w = httptest.NewRecorder()
	c, _ = gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodPost, "/api/timers/stop", nil)
	api.Timers.StopTimer(c)
	require.Equal(t, http.StatusOK, w.Code)

	service.EXPECT().StopTimer(mock.Anything, userID).Return(nil, errors.New("boom")).Once()
	w = httptest.NewRecorder()
	c, _ = gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodPost, "/api/timers/stop", nil)
	api.Timers.StopTimer(c)
	require.Equal(t, http.StatusInternalServerError, w.Code)

	service.EXPECT().CurrentTimer(mock.Anything, userID).Return(nil, nil).Once()
	w = httptest.NewRecorder()
	c, _ = gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "/api/timers/current", nil)
	api.Timers.CurrentTimer(c)
	require.Equal(t, http.StatusOK, w.Code)

	service.EXPECT().
		CurrentTimer(mock.Anything, userID).
		Return(&domain.TimerSession{ID: timerID, StartedAt: startedAt, ProblemID: &problemID}, nil).
		Once()
	w = httptest.NewRecorder()
	c, _ = gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "/api/timers/current", nil)
	api.Timers.CurrentTimer(c)
	require.Equal(t, http.StatusOK, w.Code)

	service.EXPECT().CurrentTimer(mock.Anything, userID).Return(nil, errors.New("boom")).Once()
	w = httptest.NewRecorder()
	c, _ = gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "/api/timers/current", nil)
	api.Timers.CurrentTimer(c)
	require.Equal(t, http.StatusInternalServerError, w.Code)
}
