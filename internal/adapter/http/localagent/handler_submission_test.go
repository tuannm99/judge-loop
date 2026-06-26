package localagent

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
)

func TestLocalHandlerGetSubmissionStatus(t *testing.T) {
	gin.SetMode(gin.TestMode)

	server := newTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		_ = json.NewEncoder(w).Encode(SubmissionStatusResponse{ID: "sub-1", Status: "accepted"})
	})
	defer server.Close()

	h := NewHandler(NewAPIClient(server.URL, "user-1"), "user-1", "")
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Params = gin.Params{{Key: "id", Value: "sub-1"}}
	c.Request = httptest.NewRequest(http.MethodGet, "/local/submissions/sub-1", nil)
	h.GetSubmissionStatus(c)
	require.Equal(t, http.StatusOK, w.Code)
}

func TestLocalHandlerGetSubmissionStatusReturnsGatewayError(t *testing.T) {
	gin.SetMode(gin.TestMode)

	h := NewHandler(NewAPIClient("http://127.0.0.1:1", "user-1"), "user-1", "")
	h.client.http.Timeout = 50 * time.Millisecond
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Params = gin.Params{{Key: "id", Value: "sub-1"}}
	c.Request = httptest.NewRequest(http.MethodGet, "/local/submissions/sub-1", nil)
	h.GetSubmissionStatus(c)
	require.Equal(t, http.StatusBadGateway, w.Code)
}
