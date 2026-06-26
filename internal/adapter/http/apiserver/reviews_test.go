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

func TestReviewsAPIGetReviewsToday(t *testing.T) {
	gin.SetMode(gin.TestMode)

	problemID := uuid.New()
	cases := []struct {
		name       string
		reviews    []outport.DueReview
		err        error
		wantStatus int
	}{
		{
			name:       "success",
			reviews:    []outport.DueReview{{ProblemID: problemID}},
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
			service := inmocks.NewMockReviewService(t)
			api := New(nil, nil, nil, nil, service, nil, nil, userID)

			service.EXPECT().GetReviewsToday(mock.Anything, userID).Return(tc.reviews, tc.err).Once()
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			c.Request = httptest.NewRequest(http.MethodGet, "/api/reviews/today", nil)
			api.Reviews.GetReviewsToday(c)
			require.Equal(t, tc.wantStatus, w.Code)
		})
	}
}
