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

func TestSubmissionsAPICreateSubmission(t *testing.T) {
	gin.SetMode(gin.TestMode)

	problemID := uuid.New()
	sessionID := uuid.New()
	submissionID := uuid.New()
	cases := []struct {
		name       string
		body       string
		language   string
		code       string
		sessionID  *uuid.UUID
		result     *domain.Submission
		err        error
		wantStatus int
		wantCall   bool
	}{
		{
			name: "success",
			body: `{"problem_id":"` + problemID.String() +
				`","language":"go","code":"package main{}","session_id":"` + sessionID.String() + `"}`,
			language:   "go",
			code:       "package main{}",
			sessionID:  &sessionID,
			result:     &domain.Submission{ID: submissionID, Status: domain.StatusPending},
			wantStatus: http.StatusCreated,
			wantCall:   true,
		},
		{name: "malformed body", body: "{", wantStatus: http.StatusBadRequest},
		{
			name:       "bad problem id",
			body:       `{"problem_id":"bad","language":"go","code":"x"}`,
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "unsupported language",
			body:       `{"problem_id":"` + problemID.String() + `","language":"ruby","code":"puts 1"}`,
			language:   "ruby",
			code:       "puts 1",
			err:        domain.ErrUnsupportedSubmissionLanguage,
			wantStatus: http.StatusBadRequest,
			wantCall:   true,
		},
		{
			name:       "service error with invalid session id ignored",
			body:       `{"problem_id":"` + problemID.String() + `","language":"go","code":"x","session_id":"bad"}`,
			language:   "go",
			code:       "x",
			err:        errors.New("boom"),
			wantStatus: http.StatusInternalServerError,
			wantCall:   true,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			userID := uuid.New()
			service := inmocks.NewMockSubmissionService(t)
			api := New(nil, service, nil, nil, nil, nil, nil, userID)
			if tc.wantCall {
				service.EXPECT().
					CreateSubmission(mock.Anything, userID, problemID, tc.language, tc.code, tc.sessionID).
					Return(tc.result, tc.err).
					Once()
			}

			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			c.Request = httptest.NewRequest(http.MethodPost, "/api/submissions", bytes.NewBufferString(tc.body))
			c.Request.Header.Set("Content-Type", "application/json")
			api.Submissions.CreateSubmission(c)
			require.Equal(t, tc.wantStatus, w.Code)
		})
	}
}

func TestSubmissionsAPIGetSubmission(t *testing.T) {
	gin.SetMode(gin.TestMode)

	submissionID := uuid.New()
	cases := []struct {
		name       string
		idParam    string
		result     *domain.Submission
		err        error
		wantStatus int
		wantCall   bool
	}{
		{
			name:       "success",
			idParam:    submissionID.String(),
			result:     &domain.Submission{ID: submissionID},
			wantStatus: http.StatusOK,
			wantCall:   true,
		},
		{name: "bad id", idParam: "bad", wantStatus: http.StatusBadRequest},
		{name: "not found", idParam: submissionID.String(), wantStatus: http.StatusNotFound, wantCall: true},
		{
			name:       "service error",
			idParam:    submissionID.String(),
			err:        errors.New("boom"),
			wantStatus: http.StatusInternalServerError,
			wantCall:   true,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			service := inmocks.NewMockSubmissionService(t)
			api := New(nil, service, nil, nil, nil, nil, nil, uuid.New())
			if tc.wantCall {
				service.EXPECT().GetSubmission(mock.Anything, submissionID).Return(tc.result, tc.err).Once()
			}

			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			c.Params = gin.Params{{Key: "id", Value: tc.idParam}}
			c.Request = httptest.NewRequest(http.MethodGet, "/api/submissions/"+tc.idParam, nil)
			api.Submissions.GetSubmission(c)
			require.Equal(t, tc.wantStatus, w.Code)
		})
	}
}

func TestSubmissionsAPIListSubmissions(t *testing.T) {
	gin.SetMode(gin.TestMode)

	problemID := uuid.New()
	submissionID := uuid.New()
	cases := []struct {
		name       string
		path       string
		problemID  *uuid.UUID
		result     []domain.Submission
		err        error
		wantStatus int
	}{
		{
			name:       "success",
			path:       "/api/submissions/history?problem_id=" + problemID.String(),
			problemID:  &problemID,
			result:     []domain.Submission{{ID: submissionID}},
			wantStatus: http.StatusOK,
		},
		{
			name:       "invalid problem id ignored then service error",
			path:       "/api/submissions/history?problem_id=bad",
			err:        errors.New("boom"),
			wantStatus: http.StatusInternalServerError,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			userID := uuid.New()
			service := inmocks.NewMockSubmissionService(t)
			api := New(nil, service, nil, nil, nil, nil, nil, userID)

			service.EXPECT().
				ListSubmissions(mock.Anything, userID, tc.problemID, 20, 0).
				Return(tc.result, tc.err).
				Once()
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			c.Request = httptest.NewRequest(http.MethodGet, tc.path, nil)
			api.Submissions.ListSubmissions(c)
			require.Equal(t, tc.wantStatus, w.Code)
		})
	}
}
