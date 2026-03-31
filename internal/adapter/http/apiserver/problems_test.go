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
	outport "github.com/tuannm99/judge-loop/internal/port/out"
)

func TestProblemHandlers(t *testing.T) {
	gin.SetMode(gin.TestMode)
	userID := uuid.New()
	service := inmocks.NewMockProblemService(t)
	api := New(service, nil, nil, nil, nil, nil, userID)

	t.Run("list problems success", func(t *testing.T) {
		diff := domain.DifficultyEasy
		prov := domain.ProviderLeetCode
		service.EXPECT().ListProblems(mock.Anything, mock.MatchedBy(func(f outport.ProblemFilter) bool {
			return f.Tag == "array" &&
				f.Pattern == "dp" &&
				*f.Difficulty == diff &&
				*f.Provider == prov &&
				f.Limit == 5 &&
				f.Offset == 2
		})).Return([]domain.Problem{{Slug: "two-sum"}}, 1, nil).Once()

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest(
			http.MethodGet,
			"/?difficulty=easy&tag=array&pattern=dp&provider=leetcode&limit=5&offset=2",
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
}
