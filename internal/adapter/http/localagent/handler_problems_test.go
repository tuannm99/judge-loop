package localagent

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLocalHandlerListProblems(t *testing.T) {
	gin.SetMode(gin.TestMode)

	server := newTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/api/problems", r.URL.Path)
		assert.Equal(t, "20", r.URL.Query().Get("limit"))
		_ = json.NewEncoder(w).Encode(gin.H{"problems": []gin.H{}, "total": 0})
	})
	defer server.Close()

	h := NewHandler(NewAPIClient(server.URL, "user-1"), "user-1", "")
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "/local/problems?limit=20", nil)
	h.ListProblems(c)
	require.Equal(t, http.StatusOK, w.Code)
	require.Contains(t, w.Body.String(), `"total":0`)
}

func TestLocalHandlerSuggestProblem(t *testing.T) {
	gin.SetMode(gin.TestMode)

	server := newTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/api/problems/suggest", r.URL.Path)
		_ = json.NewEncoder(w).Encode(gin.H{"id": "p1", "title": "Two Sum"})
	})
	defer server.Close()

	h := NewHandler(NewAPIClient(server.URL, "user-1"), "user-1", "")
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "/local/problems/suggest", nil)
	h.SuggestProblem(c)
	require.Equal(t, http.StatusOK, w.Code)
}

func TestLocalHandlerGetProblem(t *testing.T) {
	gin.SetMode(gin.TestMode)

	server := newTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/api/problems/two-sum", r.URL.Path)
		_ = json.NewEncoder(w).Encode(gin.H{"id": "p1", "slug": "two-sum"})
	})
	defer server.Close()

	h := NewHandler(NewAPIClient(server.URL, "user-1"), "user-1", "")
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Params = gin.Params{{Key: "id", Value: "two-sum"}}
	c.Request = httptest.NewRequest(http.MethodGet, "/local/problems/two-sum", nil)
	h.GetProblem(c)
	require.Equal(t, http.StatusOK, w.Code)
}
