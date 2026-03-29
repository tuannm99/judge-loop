package main

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// Sync handles POST /local/sync.
// Real registry sync is implemented in Milestone 7.
func (h *Handler) Sync(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"synced":  false,
		"message": "registry sync not yet implemented (Milestone 7)",
	})
}
