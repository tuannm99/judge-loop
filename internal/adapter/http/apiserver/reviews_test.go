package apiserver

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	inmocks "github.com/tuannm99/judge-loop/internal/port/in/mocks"
	outport "github.com/tuannm99/judge-loop/internal/port/out"
)

func TestReviewHandlers(t *testing.T) {
	gin.SetMode(gin.TestMode)
	userID := uuid.New()
	service := inmocks.NewMockReviewService(t)
	api := New(nil, nil, nil, nil, service, nil, nil, userID)
	problemID := uuid.New()

	service.EXPECT().GetReviewsToday(mock.Anything, userID).Return([]outport.DueReview{{ProblemID: problemID}}, nil).Once()
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "/api/reviews/today", nil)
	api.Reviews.GetReviewsToday(c)
	require.Equal(t, http.StatusOK, w.Code)

	service.EXPECT().GetReviewsToday(mock.Anything, userID).Return(nil, errors.New("boom")).Once()
	w = httptest.NewRecorder()
	c, _ = gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "/api/reviews/today", nil)
	api.Reviews.GetReviewsToday(c)
	require.Equal(t, http.StatusInternalServerError, w.Code)
}
