package apiserver

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// GetProgressToday handles GET /api/progress/today
func (h *Handler) GetProgressToday(c *gin.Context) {
	progress, err := h.Service.GetProgressToday(c.Request.Context(), h.UserID)
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
func (h *Handler) GetStreak(c *gin.Context) {
	streak, err := h.Service.GetStreak(c.Request.Context(), h.UserID)
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
