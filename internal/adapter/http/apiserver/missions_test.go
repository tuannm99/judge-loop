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
	"github.com/tuannm99/judge-loop/internal/domain"
	inmocks "github.com/tuannm99/judge-loop/internal/port/in/mocks"
)

func TestMissionsAPIGetDailyMission(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name       string
		mission    *domain.DailyMission
		err        error
		wantStatus int
		wantBody   string
	}{
		{
			name:       "returns mission",
			mission:    &domain.DailyMission{RequiredTasks: []domain.MissionTask{{Slug: "two-sum"}}},
			wantStatus: http.StatusOK,
			wantBody:   `"slug":"two-sum"`,
		},
		{
			name:       "returns empty task arrays",
			wantStatus: http.StatusOK,
			wantBody:   `"required_tasks":[]`,
		},
		{
			name:       "service error",
			err:        errors.New("boom"),
			wantStatus: http.StatusInternalServerError,
			wantBody:   `"error":"boom"`,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			userID := uuid.New()
			service := inmocks.NewMockMissionService(t)
			api := New(nil, nil, nil, nil, nil, nil, service, userID)
			service.EXPECT().
				GetDailyMission(mock.Anything, userID).
				Return(test.mission, test.err).
				Once()

			recorder := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(recorder)
			c.Request = httptest.NewRequest(http.MethodGet, "/api/missions/today", nil)
			api.Missions.GetDailyMission(c)

			require.Equal(t, test.wantStatus, recorder.Code)
			require.Contains(t, recorder.Body.String(), test.wantBody)
		})
	}
}
