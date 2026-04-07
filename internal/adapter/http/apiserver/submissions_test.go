package apiserver

import (
	"bytes"
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

func TestSubmissionHandlers(t *testing.T) {
	gin.SetMode(gin.TestMode)
	userID := uuid.New()
	service := inmocks.NewMockSubmissionService(t)
	api := New(nil, service, nil, nil, nil, nil, nil, userID)

	problemID := uuid.New()
	sessionID := uuid.New()
	submissionID := uuid.New()

	t.Run("create submission success", func(t *testing.T) {
		service.EXPECT().CreateSubmission(mock.Anything, userID, problemID, "go", "package main{}", &sessionID).
			Return(&domain.Submission{ID: submissionID, Status: domain.StatusPending}, nil).Once()

		body := `{"problem_id":"` + problemID.String() + `","language":"go","code":"package main{}","session_id":"` + sessionID.String() + `"}`
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest(http.MethodPost, "/api/submissions", bytes.NewBufferString(body))
		c.Request.Header.Set("Content-Type", "application/json")
		api.Submissions.CreateSubmission(c)
		require.Equal(t, http.StatusCreated, w.Code)
	})

	t.Run("create submission bad request and server error", func(t *testing.T) {
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest(http.MethodPost, "/api/submissions", bytes.NewBufferString("{"))
		c.Request.Header.Set("Content-Type", "application/json")
		api.Submissions.CreateSubmission(c)
		require.Equal(t, http.StatusBadRequest, w.Code)

		body := `{"problem_id":"bad","language":"go","code":"x"}`
		w = httptest.NewRecorder()
		c, _ = gin.CreateTestContext(w)
		c.Request = httptest.NewRequest(http.MethodPost, "/api/submissions", bytes.NewBufferString(body))
		c.Request.Header.Set("Content-Type", "application/json")
		api.Submissions.CreateSubmission(c)
		require.Equal(t, http.StatusBadRequest, w.Code)

		service.EXPECT().CreateSubmission(mock.Anything, userID, problemID, "go", "x", (*uuid.UUID)(nil)).
			Return(nil, errors.New("boom")).Once()
		body = `{"problem_id":"` + problemID.String() + `","language":"go","code":"x","session_id":"bad"}`
		w = httptest.NewRecorder()
		c, _ = gin.CreateTestContext(w)
		c.Request = httptest.NewRequest(http.MethodPost, "/api/submissions", bytes.NewBufferString(body))
		c.Request.Header.Set("Content-Type", "application/json")
		api.Submissions.CreateSubmission(c)
		require.Equal(t, http.StatusInternalServerError, w.Code)
	})

	t.Run("get submission and list submissions branches", func(t *testing.T) {
		service.EXPECT().
			GetSubmission(mock.Anything, submissionID).
			Return(&domain.Submission{ID: submissionID}, nil).
			Once()
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Params = gin.Params{{Key: "id", Value: submissionID.String()}}
		c.Request = httptest.NewRequest(http.MethodGet, "/api/submissions/"+submissionID.String(), nil)
		api.Submissions.GetSubmission(c)
		require.Equal(t, http.StatusOK, w.Code)

		w = httptest.NewRecorder()
		c, _ = gin.CreateTestContext(w)
		c.Params = gin.Params{{Key: "id", Value: "bad"}}
		c.Request = httptest.NewRequest(http.MethodGet, "/api/submissions/bad", nil)
		api.Submissions.GetSubmission(c)
		require.Equal(t, http.StatusBadRequest, w.Code)

		service.EXPECT().GetSubmission(mock.Anything, submissionID).Return(nil, nil).Once()
		w = httptest.NewRecorder()
		c, _ = gin.CreateTestContext(w)
		c.Params = gin.Params{{Key: "id", Value: submissionID.String()}}
		c.Request = httptest.NewRequest(http.MethodGet, "/api/submissions/"+submissionID.String(), nil)
		api.Submissions.GetSubmission(c)
		require.Equal(t, http.StatusNotFound, w.Code)

		service.EXPECT().GetSubmission(mock.Anything, submissionID).Return(nil, errors.New("boom")).Once()
		w = httptest.NewRecorder()
		c, _ = gin.CreateTestContext(w)
		c.Params = gin.Params{{Key: "id", Value: submissionID.String()}}
		c.Request = httptest.NewRequest(http.MethodGet, "/api/submissions/"+submissionID.String(), nil)
		api.Submissions.GetSubmission(c)
		require.Equal(t, http.StatusInternalServerError, w.Code)

		service.EXPECT().
			ListSubmissions(mock.Anything, userID, &problemID, 20, 0).
			Return([]domain.Submission{{ID: submissionID}}, nil).
			Once()
		w = httptest.NewRecorder()
		c, _ = gin.CreateTestContext(w)
		c.Request = httptest.NewRequest(http.MethodGet, "/api/submissions/history?problem_id="+problemID.String(), nil)
		api.Submissions.ListSubmissions(c)
		require.Equal(t, http.StatusOK, w.Code)

		service.EXPECT().
			ListSubmissions(mock.Anything, userID, (*uuid.UUID)(nil), 20, 0).
			Return(nil, errors.New("boom")).
			Once()
		w = httptest.NewRecorder()
		c, _ = gin.CreateTestContext(w)
		c.Request = httptest.NewRequest(http.MethodGet, "/api/submissions/history?problem_id=bad", nil)
		api.Submissions.ListSubmissions(c)
		require.Equal(t, http.StatusInternalServerError, w.Code)
	})
}
