package apiserver

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// GetDailyMission handles GET /api/missions/today
func (h *MissionsAPI) GetDailyMission(c *gin.Context) {
	mission, err := h.service.GetDailyMission(c.Request.Context(), h.userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if mission == nil {
		c.JSON(http.StatusOK, gin.H{
			"required_tasks": []any{},
			"review_tasks":   []any{},
			"optional_tasks": []any{},
		})
		return
	}
	c.JSON(http.StatusOK, mission)
}
