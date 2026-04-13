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

func TestProblemHandlers(t *testing.T) {
	gin.SetMode(gin.TestMode)
	userID := uuid.New()
	service := inmocks.NewMockProblemService(t)
	api := New(service, nil, nil, nil, nil, nil, nil, userID)

	t.Run("list problems success", func(t *testing.T) {
		diff := domain.DifficultyEasy
		prov := domain.ProviderLeetCode
		service.EXPECT().ListProblems(mock.Anything, mock.MatchedBy(func(f outport.ProblemFilter) bool {
			return len(f.Tags) == 4 &&
				f.Tags[0] == "array" &&
				f.Tags[1] == "hash-table" &&
				f.Tags[2] == "dp" &&
				f.Tags[3] == "graph" &&
				*f.Difficulty == diff &&
				*f.Provider == prov &&
				f.Limit == 5 &&
				f.Offset == 2
		})).Return([]domain.Problem{{Slug: "two-sum"}}, 1, nil).Once()

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest(
			http.MethodGet,
			"/?difficulty=easy&tag=array&tags=hash-table&pattern=dp&patterns=graph&provider=leetcode&limit=5&offset=2",
			nil,
		)
		api.Problems.ListProblems(c)
		require.Equal(t, http.StatusOK, w.Code)
		require.Contains(t, w.Body.String(), `"total":1`)
	})

	t.Run("list problems error and invalid pagination ignored", func(t *testing.T) {
		service.EXPECT().ListProblems(mock.Anything, mock.MatchedBy(func(f outport.ProblemFilter) bool {
			return f.Limit == 0 && f.Offset == 0
		})).Return(nil, 0, errors.New("boom")).Once()

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest(http.MethodGet, "/?limit=x&offset=-1", nil)
		api.Problems.ListProblems(c)
		require.Equal(t, http.StatusInternalServerError, w.Code)
	})

	t.Run("list problem labels", func(t *testing.T) {
		service.EXPECT().ListProblemLabels(mock.Anything).Return([]string{"array"}, nil).Once()

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest(http.MethodGet, "/api/problems/labels", nil)
		api.Problems.ListProblemLabels(c)
		require.Equal(t, http.StatusOK, w.Code)
		require.Contains(t, w.Body.String(), `"tags":["array"]`)

		service.EXPECT().ListProblemLabels(mock.Anything).Return(nil, errors.New("boom")).Once()
		w = httptest.NewRecorder()
		c, _ = gin.CreateTestContext(w)
		c.Request = httptest.NewRequest(http.MethodGet, "/api/problems/labels", nil)
		api.Problems.ListProblemLabels(c)
		require.Equal(t, http.StatusInternalServerError, w.Code)
	})

	t.Run("problem label crud", func(t *testing.T) {
		labelID := uuid.New()

		service.EXPECT().ListProblemLabelRecords(mock.Anything, "tag").
			Return([]domain.ProblemLabel{{ID: labelID, Kind: "tag", Slug: "array", Name: "Array"}}, nil).Once()
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest(http.MethodGet, "/api/problem-labels?kind=tag", nil)
		q := c.Request.URL.Query()
		q.Set("kind", "tag")
		c.Request.URL.RawQuery = q.Encode()
		api.Problems.ListProblemLabelRecords(c)
		require.Equal(t, http.StatusOK, w.Code)
		require.Contains(t, w.Body.String(), `"slug":"array"`)

		body, err := json.Marshal(problemLabelRequest{Kind: "pattern", Slug: "sliding-window", Name: "Sliding Window"})
		require.NoError(t, err)
		service.EXPECT().CreateProblemLabel(mock.Anything, "tag", "sliding-window", "Sliding Window").
			Return(&domain.ProblemLabel{ID: labelID, Kind: "tag", Slug: "sliding-window", Name: "Sliding Window"}, nil).Once()
		w = httptest.NewRecorder()
		c, _ = gin.CreateTestContext(w)
		c.Request = httptest.NewRequest(http.MethodPost, "/api/problem-labels", bytes.NewReader(body))
		c.Request.Header.Set("Content-Type", "application/json")
		api.Problems.CreateProblemLabel(c)
		require.Equal(t, http.StatusCreated, w.Code)

		updateBody, err := json.Marshal(updateProblemLabelRequest{Slug: "two-pointers", Name: "Two Pointers"})
		require.NoError(t, err)
		service.EXPECT().UpdateProblemLabel(mock.Anything, labelID, "two-pointers", "Two Pointers").
			Return(&domain.ProblemLabel{ID: labelID, Kind: "tag", Slug: "two-pointers", Name: "Two Pointers"}, nil).Once()
		w = httptest.NewRecorder()
		c, _ = gin.CreateTestContext(w)
		c.Params = gin.Params{{Key: "id", Value: labelID.String()}}
		c.Request = httptest.NewRequest(http.MethodPut, "/api/problem-labels/"+labelID.String(), bytes.NewReader(updateBody))
		c.Request.Header.Set("Content-Type", "application/json")
		api.Problems.UpdateProblemLabel(c)
		require.Equal(t, http.StatusOK, w.Code)

		service.EXPECT().DeleteProblemLabel(mock.Anything, labelID).Return(nil).Once()
		w = httptest.NewRecorder()
		c, _ = gin.CreateTestContext(w)
		c.Params = gin.Params{{Key: "id", Value: labelID.String()}}
		c.Request = httptest.NewRequest(http.MethodDelete, "/api/problem-labels/"+labelID.String(), nil)
		api.Problems.DeleteProblemLabel(c)
		require.Equal(t, http.StatusNoContent, w.Code)
	})

	t.Run("get problem branches", func(t *testing.T) {
		service.EXPECT().GetProblem(mock.Anything, "slug").Return(&domain.Problem{Slug: "slug"}, nil).Once()
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Params = gin.Params{{Key: "id", Value: "slug"}}
		c.Request = httptest.NewRequest(http.MethodGet, "/api/problems/slug", nil)
		api.Problems.GetProblem(c)
		require.Equal(t, http.StatusOK, w.Code)

		service.EXPECT().GetProblem(mock.Anything, "missing").Return(nil, nil).Once()
		w = httptest.NewRecorder()
		c, _ = gin.CreateTestContext(w)
		c.Params = gin.Params{{Key: "id", Value: "missing"}}
		c.Request = httptest.NewRequest(http.MethodGet, "/api/problems/missing", nil)
		api.Problems.GetProblem(c)
		require.Equal(t, http.StatusNotFound, w.Code)

		service.EXPECT().GetProblem(mock.Anything, "err").Return(nil, errors.New("boom")).Once()
		w = httptest.NewRecorder()
		c, _ = gin.CreateTestContext(w)
		c.Params = gin.Params{{Key: "id", Value: "err"}}
		c.Request = httptest.NewRequest(http.MethodGet, "/api/problems/err", nil)
		api.Problems.GetProblem(c)
		require.Equal(t, http.StatusInternalServerError, w.Code)
	})

	t.Run("get problem test cases", func(t *testing.T) {
		problemID := uuid.New()
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
	})

	t.Run("suggest problem branches", func(t *testing.T) {
		service.EXPECT().SuggestProblem(mock.Anything, userID).Return(&domain.Problem{Slug: "x"}, nil).Once()
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest(http.MethodGet, "/api/problems/suggest", nil)
		api.Problems.SuggestProblem(c)
		require.Equal(t, http.StatusOK, w.Code)

		service.EXPECT().SuggestProblem(mock.Anything, userID).Return(nil, nil).Once()
		w = httptest.NewRecorder()
		c, _ = gin.CreateTestContext(w)
		c.Request = httptest.NewRequest(http.MethodGet, "/api/problems/suggest", nil)
		api.Problems.SuggestProblem(c)
		require.Equal(t, http.StatusNotFound, w.Code)

		service.EXPECT().SuggestProblem(mock.Anything, userID).Return(nil, errors.New("boom")).Once()
		w = httptest.NewRecorder()
		c, _ = gin.CreateTestContext(w)
		c.Request = httptest.NewRequest(http.MethodGet, "/api/problems/suggest", nil)
		api.Problems.SuggestProblem(c)
		require.Equal(t, http.StatusInternalServerError, w.Code)
	})

	t.Run("update problem branches", func(t *testing.T) {
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

		service.EXPECT().UpdateProblemWithTestCases(mock.Anything, problemID, domain.ProblemManifest{
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
		}, []domain.TestCase{}).Return(&domain.Problem{ID: problemID, Slug: "two-sum"}, nil).Once()

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Params = gin.Params{{Key: "id", Value: problemID.String()}}
		c.Request = httptest.NewRequest(http.MethodPut, "/api/problems/"+problemID.String(), bytes.NewReader(body))
		c.Request.Header.Set("Content-Type", "application/json")
		api.Problems.UpdateProblem(c)
		require.Equal(t, http.StatusOK, w.Code)

		w = httptest.NewRecorder()
		c, _ = gin.CreateTestContext(w)
		c.Params = gin.Params{{Key: "id", Value: "bad"}}
		c.Request = httptest.NewRequest(http.MethodPut, "/api/problems/bad", bytes.NewReader(body))
		c.Request.Header.Set("Content-Type", "application/json")
		api.Problems.UpdateProblem(c)
		require.Equal(t, http.StatusBadRequest, w.Code)

		service.EXPECT().UpdateProblemWithTestCases(mock.Anything, problemID, mock.Anything, []domain.TestCase{}).Return(nil, errors.New("boom")).Once()
		w = httptest.NewRecorder()
		c, _ = gin.CreateTestContext(w)
		c.Params = gin.Params{{Key: "id", Value: problemID.String()}}
		c.Request = httptest.NewRequest(http.MethodPut, "/api/problems/"+problemID.String(), bytes.NewReader(body))
		c.Request.Header.Set("Content-Type", "application/json")
		api.Problems.UpdateProblem(c)
		require.Equal(t, http.StatusInternalServerError, w.Code)
	})

	t.Run("contribute problem branches", func(t *testing.T) {
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

		service.EXPECT().ContributeProblem(mock.Anything, domain.ProblemManifest{
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
			Version: 1,
		}, []domain.TestCase{
			{Input: "1 2", Expected: "3", OrderIdx: 0},
			{Input: "2 3", Expected: "5", IsHidden: true, OrderIdx: 1},
		}).Return(&domain.Problem{
			Slug: "two-sum",
			StarterCode: map[string]string{
				"python": "class Solution:\n    pass\n",
			},
		}, nil).Once()

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest(http.MethodPost, "/api/problems/contribute", bytes.NewReader(body))
		c.Request.Header.Set("Content-Type", "application/json")
		api.Problems.ContributeProblem(c)
		require.Equal(t, http.StatusCreated, w.Code)

		w = httptest.NewRecorder()
		c, _ = gin.CreateTestContext(w)
		c.Request = httptest.NewRequest(http.MethodPost, "/api/problems/contribute", bytes.NewBufferString("{"))
		c.Request.Header.Set("Content-Type", "application/json")
		api.Problems.ContributeProblem(c)
		require.Equal(t, http.StatusBadRequest, w.Code)

		service.EXPECT().ContributeProblem(mock.Anything, mock.Anything, mock.Anything).Return(nil, errors.New("boom")).Once()
		w = httptest.NewRecorder()
		c, _ = gin.CreateTestContext(w)
		c.Request = httptest.NewRequest(http.MethodPost, "/api/problems/contribute", bytes.NewReader(body))
		c.Request.Header.Set("Content-Type", "application/json")
		api.Problems.ContributeProblem(c)
		require.Equal(t, http.StatusInternalServerError, w.Code)
	})
}
