package main

import (
	"context"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

// StatusToday handles GET /local/status/today.
// It tries the api-server first; if unreachable it falls back to local timer state.
func (h *Handler) StatusToday(c *gin.Context) {
	ctx, cancel := context.WithTimeout(c.Request.Context(), 2*time.Second)
	defer cancel()

	active := h.timer.Active() != nil

	progress, err := h.client.ProgressToday(ctx)
	if err != nil {
		// server unreachable — derive from local state only
		practiced := active
		msg := "No practice yet today. Start a session!"
		if practiced {
			msg = "Timer is running. Keep going!"
		}
		c.JSON(http.StatusOK, gin.H{
			"practiced":    practiced,
			"solved_count": 0,
			"active_timer": active,
			"message":      msg,
			"server_error": err.Error(),
		})
		return
	}

	practiced := progress.Solved > 0 || active
	msg := buildStatusMessage(progress.Solved, active)

	c.JSON(http.StatusOK, gin.H{
		"practiced":    practiced,
		"solved_count": progress.Solved,
		"active_timer": active,
		"streak":       progress.Streak,
		"message":      msg,
	})
}

func buildStatusMessage(solved int, active bool) string {
	switch {
	case solved > 0 && active:
		return "Timer running — keep it up!"
	case solved > 0:
		return "Good work today! Come back tomorrow."
	case active:
		return "Timer is running. Submit when ready."
	default:
		return "No practice yet today. Start a session!"
	}
}
