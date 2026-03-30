package main

import (
	"context"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

// GetSubmissionStatus proxies GET /local/submissions/:id to GET /api/submissions/:id.
// The Neovim plugin polls this endpoint until a terminal verdict arrives.
func (h *Handler) GetSubmissionStatus(c *gin.Context) {
	id := c.Param("id")

	ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
	defer cancel()

	resp, err := h.client.GetSubmission(ctx, id)
	if err != nil {
		c.JSON(http.StatusBadGateway, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, resp)
}
