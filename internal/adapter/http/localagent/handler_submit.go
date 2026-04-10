package localagent

import (
	"context"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

type submitRequest struct {
	ProblemID string `json:"problem_id" binding:"required"`
	Language  string `json:"language"   binding:"required"`
	Code      string `json:"code"       binding:"required"`
}

// Submit handles POST /local/submit.
// Attaches any active local timer session, then proxies to api-server.
// Requires the server to be reachable — submissions cannot be queued offline.
func (h *Handler) Submit(c *gin.Context) {
	var req submitRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	apiReq := SubmitRequest{
		ProblemID: req.ProblemID,
		Language:  req.Language,
		Code:      req.Code,
	}

	// attach active timer session if one is running
	if active := h.timer.Active(); active != nil && active.ServerID != nil {
		apiReq.SessionID = active.ServerID.String()
	}

	ctx, cancel := context.WithTimeout(c.Request.Context(), 10*time.Second)
	defer cancel()

	resp, err := h.client.Submit(ctx, apiReq)
	if err != nil {
		c.JSON(http.StatusBadGateway, gin.H{"error": "api-server unreachable: " + err.Error()})
		return
	}

	c.JSON(http.StatusCreated, resp)
}
