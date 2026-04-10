package localagent

import (
	"context"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

// ListProblems proxies GET /local/problems to GET /api/problems.
func (h *Handler) ListProblems(c *gin.Context) {
	h.proxyProblemGet(c, "/api/problems")
}

// SuggestProblem proxies GET /local/problems/suggest to GET /api/problems/suggest.
func (h *Handler) SuggestProblem(c *gin.Context) {
	h.proxyProblemGet(c, "/api/problems/suggest")
}

// GetProblem proxies GET /local/problems/:id to GET /api/problems/:id.
func (h *Handler) GetProblem(c *gin.Context) {
	h.proxyProblemGet(c, "/api/problems/"+c.Param("id"))
}

func (h *Handler) proxyProblemGet(c *gin.Context, path string) {
	ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
	defer cancel()

	raw, err := h.client.ProxyGet(ctx, path, c.Request.URL.Query())
	if err != nil {
		c.JSON(http.StatusBadGateway, gin.H{"error": err.Error()})
		return
	}
	c.Data(http.StatusOK, "application/json; charset=utf-8", raw)
}
