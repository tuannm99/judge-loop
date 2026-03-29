package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// GetProgressToday handles GET /api/progress/today
func (h *Handler) GetProgressToday(c *gin.Context) {
	ds, err := h.Sessions.GetOrCreateToday(c.Request.Context(), h.UserID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	streak, err := h.Sessions.GetStreak(c.Request.Context(), h.UserID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"date":                ds.Date.Format("2006-01-02"),
		"solved":              ds.SolvedCount,
		"attempted":           ds.AttemptedCount,
		"time_spent_minutes":  ds.TimeSpentSecs / 60,
		"streak":              streak.Current,
	})
}

// GetStreak handles GET /api/streak
func (h *Handler) GetStreak(c *gin.Context) {
	streak, err := h.Sessions.GetStreak(c.Request.Context(), h.UserID)
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
