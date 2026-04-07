package apiserver

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
	inmocks "github.com/tuannm99/judge-loop/internal/port/in/mocks"
)

func TestNewRouterHealth(t *testing.T) {
	router := NewRouter(New(
		inmocks.NewMockProblemService(t),
		inmocks.NewMockSubmissionService(t),
		inmocks.NewMockProgressService(t),
		inmocks.NewMockTimerService(t),
		inmocks.NewMockReviewService(t),
		inmocks.NewMockRegistryService(t),
		inmocks.NewMockMissionService(t),
		uuid.New(),
	))

	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	require.Equal(t, http.StatusOK, rec.Code)
	require.JSONEq(t, `{"status":"ok"}`, rec.Body.String())
}
