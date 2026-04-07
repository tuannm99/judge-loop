package apiserver

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// GetProgressToday handles GET /api/progress/today
func (h *ProgressAPI) GetProgressToday(c *gin.Context) {
	progress, err := h.service.GetProgressToday(c.Request.Context(), h.userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"date":               progress.Date,
		"solved":             progress.Solved,
		"attempted":          progress.Attempted,
		"time_spent_minutes": progress.TimeSpentMinutes,
		"streak":             progress.Streak,
	})
}

// GetStreak handles GET /api/streak
func (h *ProgressAPI) GetStreak(c *gin.Context) {
	streak, err := h.service.GetStreak(c.Request.Context(), h.userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"current":        streak.Current,
		"longest":        streak.Longest,
		"last_practiced": streak.LastPracticed,
	})
}

// GetGoalProgress handles GET /api/progress/goal
func (h *ProgressAPI) GetGoalProgress(c *gin.Context) {
	goal, err := h.service.GetGoalProgress(c.Request.Context(), h.userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, goal)
}
