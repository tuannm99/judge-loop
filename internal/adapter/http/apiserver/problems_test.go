package apiserver

import (
	"bytes"
	"encoding/json"
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
	outport "github.com/tuannm99/judge-loop/internal/port/out"
)

func TestProblemsAPIListProblems(t *testing.T) {
	gin.SetMode(gin.TestMode)

	cases := []struct {
		name       string
		path       string
		setup      func(*inmocks.MockProblemService)
		wantStatus int
		wantBody   string
	}{
		{
			name: "success",
			path: "/?title=%20two%20sum%20&difficulty=easy&tag=array&tags=hash-table&pattern=dp&patterns=graph&provider=leetcode&limit=5&offset=2",
			setup: func(service *inmocks.MockProblemService) {
				diff := domain.DifficultyEasy
				prov := domain.ProviderLeetCode
				service.EXPECT().ListProblems(mock.Anything, mock.MatchedBy(func(f outport.ProblemFilter) bool {
					return len(f.Tags) == 4 &&
						f.Title == "two sum" &&
						f.Tags[0] == "array" &&
						f.Tags[1] == "hash-table" &&
						f.Tags[2] == "dp" &&
						f.Tags[3] == "graph" &&
						*f.Difficulty == diff &&
						*f.Provider == prov &&
						f.Limit == 5 &&
						f.Offset == 2
				})).Return([]domain.Problem{{Slug: "two-sum"}}, 1, nil).Once()
			},
			wantStatus: http.StatusOK,
			wantBody:   `"total":1`,
		},
		{
			name: "service error and invalid pagination ignored",
			path: "/?limit=x&offset=-1",
			setup: func(service *inmocks.MockProblemService) {
				service.EXPECT().ListProblems(mock.Anything, mock.MatchedBy(func(f outport.ProblemFilter) bool {
					return f.Limit == 0 && f.Offset == 0
				})).Return(nil, 0, errors.New("boom")).Once()
			},
			wantStatus: http.StatusInternalServerError,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			service := inmocks.NewMockProblemService(t)
			api := New(service, nil, nil, nil, nil, nil, nil, uuid.New())
			tc.setup(service)

			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			c.Request = httptest.NewRequest(http.MethodGet, tc.path, nil)
			api.Problems.ListProblems(c)
			require.Equal(t, tc.wantStatus, w.Code)
			if tc.wantBody != "" {
				require.Contains(t, w.Body.String(), tc.wantBody)
			}
		})
	}
}

func TestProblemsAPIListProblemLabels(t *testing.T) {
	gin.SetMode(gin.TestMode)

	cases := []struct {
		name       string
		labels     []string
		err        error
		wantStatus int
		wantBody   string
	}{
		{name: "success", labels: []string{"array"}, wantStatus: http.StatusOK, wantBody: `"tags":["array"]`},
		{name: "service error", err: errors.New("boom"), wantStatus: http.StatusInternalServerError},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			service := inmocks.NewMockProblemService(t)
			api := New(service, nil, nil, nil, nil, nil, nil, uuid.New())

			service.EXPECT().ListProblemLabels(mock.Anything).Return(tc.labels, tc.err).Once()
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			c.Request = httptest.NewRequest(http.MethodGet, "/api/problems/labels", nil)
			api.Problems.ListProblemLabels(c)
			require.Equal(t, tc.wantStatus, w.Code)
			if tc.wantBody != "" {
				require.Contains(t, w.Body.String(), tc.wantBody)
			}
		})
	}
}

func TestProblemsAPIListProblemLabelRecords(t *testing.T) {
	gin.SetMode(gin.TestMode)

	labelID := uuid.New()
	service := inmocks.NewMockProblemService(t)
	api := New(service, nil, nil, nil, nil, nil, nil, uuid.New())

	service.EXPECT().
		ListProblemLabelRecords(mock.Anything, "tag").
		Return([]domain.ProblemLabel{{ID: labelID, Kind: "tag", Slug: "array", Name: "Array"}}, nil).
		Once()
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "/api/problem-labels?kind=tag", nil)
	api.Problems.ListProblemLabelRecords(c)
	require.Equal(t, http.StatusOK, w.Code)
	require.Contains(t, w.Body.String(), `"slug":"array"`)
}

func TestProblemsAPICreateProblemLabel(t *testing.T) {
	gin.SetMode(gin.TestMode)

	labelID := uuid.New()
	body, err := json.Marshal(problemLabelRequest{Kind: "pattern", Slug: "sliding-window", Name: "Sliding Window"})
	require.NoError(t, err)
	service := inmocks.NewMockProblemService(t)
	api := New(service, nil, nil, nil, nil, nil, nil, uuid.New())

	service.EXPECT().CreateProblemLabel(mock.Anything, "tag", "sliding-window", "Sliding Window").
		Return(&domain.ProblemLabel{ID: labelID, Kind: "tag", Slug: "sliding-window", Name: "Sliding Window"}, nil).
		Once()
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodPost, "/api/problem-labels", bytes.NewReader(body))
	c.Request.Header.Set("Content-Type", "application/json")
	api.Problems.CreateProblemLabel(c)
	require.Equal(t, http.StatusCreated, w.Code)
}

func TestProblemsAPIUpdateProblemLabel(t *testing.T) {
	gin.SetMode(gin.TestMode)

	labelID := uuid.New()
	body, err := json.Marshal(updateProblemLabelRequest{Slug: "two-pointers", Name: "Two Pointers"})
	require.NoError(t, err)
	service := inmocks.NewMockProblemService(t)
	api := New(service, nil, nil, nil, nil, nil, nil, uuid.New())

	service.EXPECT().UpdateProblemLabel(mock.Anything, labelID, "two-pointers", "Two Pointers").
		Return(&domain.ProblemLabel{ID: labelID, Kind: "tag", Slug: "two-pointers", Name: "Two Pointers"}, nil).
		Once()
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Params = gin.Params{{Key: "id", Value: labelID.String()}}
	c.Request = httptest.NewRequest(http.MethodPut, "/api/problem-labels/"+labelID.String(), bytes.NewReader(body))
	c.Request.Header.Set("Content-Type", "application/json")
	api.Problems.UpdateProblemLabel(c)
	require.Equal(t, http.StatusOK, w.Code)
}

func TestProblemsAPIDeleteProblemLabel(t *testing.T) {
	gin.SetMode(gin.TestMode)

	labelID := uuid.New()
	service := inmocks.NewMockProblemService(t)
	api := New(service, nil, nil, nil, nil, nil, nil, uuid.New())

	service.EXPECT().DeleteProblemLabel(mock.Anything, labelID).Return(nil).Once()
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Params = gin.Params{{Key: "id", Value: labelID.String()}}
	c.Request = httptest.NewRequest(http.MethodDelete, "/api/problem-labels/"+labelID.String(), nil)
	api.Problems.DeleteProblemLabel(c)
	require.Equal(t, http.StatusNoContent, w.Code)
}

func TestProblemsAPIGetProblem(t *testing.T) {
	gin.SetMode(gin.TestMode)

	cases := []struct {
		name       string
		id         string
		problem    *domain.Problem
		err        error
		wantStatus int
	}{
		{name: "success", id: "slug", problem: &domain.Problem{Slug: "slug"}, wantStatus: http.StatusOK},
		{name: "not found", id: "missing", wantStatus: http.StatusNotFound},
		{name: "service error", id: "err", err: errors.New("boom"), wantStatus: http.StatusInternalServerError},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			service := inmocks.NewMockProblemService(t)
			api := New(service, nil, nil, nil, nil, nil, nil, uuid.New())

			service.EXPECT().GetProblem(mock.Anything, tc.id).Return(tc.problem, tc.err).Once()
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			c.Params = gin.Params{{Key: "id", Value: tc.id}}
			c.Request = httptest.NewRequest(http.MethodGet, "/api/problems/"+tc.id, nil)
			api.Problems.GetProblem(c)
			require.Equal(t, tc.wantStatus, w.Code)
		})
	}
}

func TestProblemsAPIGetProblemTestCases(t *testing.T) {
	gin.SetMode(gin.TestMode)

	problemID := uuid.New()
	service := inmocks.NewMockProblemService(t)
	api := New(service, nil, nil, nil, nil, nil, nil, uuid.New())

	service.EXPECT().GetProblemTestCases(mock.Anything, problemID).Return([]domain.TestCase{
		{Input: "1 2", Expected: "3", IsHidden: true},
	}, nil).Once()
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Params = gin.Params{{Key: "id", Value: problemID.String()}}
	c.Request = httptest.NewRequest(http.MethodGet, "/api/problems/"+problemID.String()+"/test-cases", nil)
	api.Problems.GetProblemTestCases(c)
	require.Equal(t, http.StatusOK, w.Code)
	require.Contains(t, w.Body.String(), `"test_cases"`)
}

func TestProblemsAPISuggestProblem(t *testing.T) {
	gin.SetMode(gin.TestMode)

	cases := []struct {
		name       string
		problem    *domain.Problem
		err        error
		wantStatus int
	}{
		{name: "success", problem: &domain.Problem{Slug: "x"}, wantStatus: http.StatusOK},
		{name: "not found", wantStatus: http.StatusNotFound},
		{name: "service error", err: errors.New("boom"), wantStatus: http.StatusInternalServerError},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			userID := uuid.New()
			service := inmocks.NewMockProblemService(t)
			api := New(service, nil, nil, nil, nil, nil, nil, userID)

			service.EXPECT().SuggestProblem(mock.Anything, userID).Return(tc.problem, tc.err).Once()
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			c.Request = httptest.NewRequest(http.MethodGet, "/api/problems/suggest", nil)
			api.Problems.SuggestProblem(c)
			require.Equal(t, tc.wantStatus, w.Code)
		})
	}
}

func TestProblemsAPIUpdateProblem(t *testing.T) {
	gin.SetMode(gin.TestMode)

	problemID := uuid.New()
	body, err := json.Marshal(updateProblemRequest{
		Provider:            domain.ProviderLeetCode,
		ExternalID:          "1",
		Slug:                "two-sum",
		Title:               "Two Sum",
		Difficulty:          domain.DifficultyEasy,
		Tags:                []string{"array"},
		LegacyPatternTags:   []string{"hash-map"},
		SourceURL:           "https://leetcode.com/problems/two-sum",
		EstimatedTime:       15,
		DescriptionMarkdown: "## Two Sum\n\nReturn the matching pair.",
		StarterCode: map[string]string{
			"python": "class Solution:\n    pass\n",
		},
		TestCases: []contributeTestCaseRequest{},
	})
	require.NoError(t, err)
	wantManifest := domain.ProblemManifest{
		Provider:            domain.ProviderLeetCode,
		ExternalID:          "1",
		Slug:                "two-sum",
		Title:               "Two Sum",
		Difficulty:          domain.DifficultyEasy,
		Tags:                []string{"array", "hash-map"},
		SourceURL:           "https://leetcode.com/problems/two-sum",
		EstimatedTime:       15,
		DescriptionMarkdown: "## Two Sum\n\nReturn the matching pair.",
		StarterCode: map[string]string{
			"python": "class Solution:\n    pass\n",
		},
	}
	cases := []struct {
		name       string
		idParam    string
		err        error
		wantStatus int
		wantCall   bool
	}{
		{name: "success", idParam: problemID.String(), wantStatus: http.StatusOK, wantCall: true},
		{name: "bad id", idParam: "bad", wantStatus: http.StatusBadRequest},
		{
			name:       "service error",
			idParam:    problemID.String(),
			err:        errors.New("boom"),
			wantStatus: http.StatusInternalServerError,
			wantCall:   true,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			service := inmocks.NewMockProblemService(t)
			api := New(service, nil, nil, nil, nil, nil, nil, uuid.New())
			if tc.wantCall {
				service.EXPECT().
					UpdateProblemWithTestCases(mock.Anything, problemID, wantManifest, []domain.TestCase{}).
					Return(&domain.Problem{ID: problemID, Slug: "two-sum"}, tc.err).
					Once()
			}

			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			c.Params = gin.Params{{Key: "id", Value: tc.idParam}}
			c.Request = httptest.NewRequest(http.MethodPut, "/api/problems/"+tc.idParam, bytes.NewReader(body))
			c.Request.Header.Set("Content-Type", "application/json")
			api.Problems.UpdateProblem(c)
			require.Equal(t, tc.wantStatus, w.Code)
		})
	}
}

func TestProblemsAPIContributeProblem(t *testing.T) {
	gin.SetMode(gin.TestMode)

	body, err := json.Marshal(contributeProblemRequest{
		Provider:            domain.ProviderLeetCode,
		ExternalID:          "1",
		Slug:                "two-sum",
		Title:               "Two Sum",
		Difficulty:          domain.DifficultyEasy,
		Tags:                []string{"array"},
		LegacyPatternTags:   []string{"lookup"},
		SourceURL:           "https://leetcode.com/problems/two-sum",
		EstimatedTime:       15,
		DescriptionMarkdown: "## Two Sum\n\nReturn the matching pair.",
		StarterCode: map[string]string{
			"python": "class Solution:\n    pass\n",
			"go":     "package main\n\nfunc main() {}\n",
		},
		Version: 1,
		TestCases: []contributeTestCaseRequest{
			{Input: "1 2", Expected: "3"},
			{Input: "2 3", Expected: "5", IsHidden: true},
		},
	})
	require.NoError(t, err)
	wantManifest := domain.ProblemManifest{
		Provider:            domain.ProviderLeetCode,
		ExternalID:          "1",
		Slug:                "two-sum",
		Title:               "Two Sum",
		Difficulty:          domain.DifficultyEasy,
		Tags:                []string{"array", "lookup"},
		SourceURL:           "https://leetcode.com/problems/two-sum",
		EstimatedTime:       15,
		DescriptionMarkdown: "## Two Sum\n\nReturn the matching pair.",
		StarterCode: map[string]string{
			"python": "class Solution:\n    pass\n",
			"go":     "package main\n\nfunc main() {}\n",
		},
		JudgeReady: true,
		Version:    1,
	}
	wantCases := []domain.TestCase{
		{Input: "1 2", Expected: "3", OrderIdx: 0},
		{Input: "2 3", Expected: "5", IsHidden: true, OrderIdx: 1},
	}
	cases := []struct {
		name       string
		body       []byte
		err        error
		wantStatus int
		wantCall   bool
	}{
		{name: "success", body: body, wantStatus: http.StatusCreated, wantCall: true},
		{name: "bad request", body: []byte("{"), wantStatus: http.StatusBadRequest},
		{
			name:       "service error",
			body:       body,
			err:        errors.New("boom"),
			wantStatus: http.StatusInternalServerError,
			wantCall:   true,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			service := inmocks.NewMockProblemService(t)
			api := New(service, nil, nil, nil, nil, nil, nil, uuid.New())
			if tc.wantCall {
				service.EXPECT().
					ContributeProblem(mock.Anything, wantManifest, wantCases).
					Return(&domain.Problem{
						Slug: "two-sum",
						StarterCode: map[string]string{
							"python": "class Solution:\n    pass\n",
						},
					}, tc.err).
					Once()
			}

			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			c.Request = httptest.NewRequest(http.MethodPost, "/api/problems/contribute", bytes.NewReader(tc.body))
			c.Request.Header.Set("Content-Type", "application/json")
			api.Problems.ContributeProblem(c)
			require.Equal(t, tc.wantStatus, w.Code)
		})
	}
}
