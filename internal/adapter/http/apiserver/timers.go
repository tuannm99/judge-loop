package apiserver

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	postgres "github.com/tuannm99/judge-loop/internal/infrastructure/postgres"
)

// startTimerRequest is the POST /api/timers/start request body.
type startTimerRequest struct {
	ProblemID string `json:"problem_id"` // optional
}

// StartTimer handles POST /api/timers/start
func (h *TimersAPI) StartTimer(c *gin.Context) {
	var req startTimerRequest
	_ = c.ShouldBindJSON(&req) // optional body

	var problemID *uuid.UUID
	if req.ProblemID != "" {
		if id, err := uuid.Parse(req.ProblemID); err == nil {
			problemID = &id
		}
	}

	ts, err := h.service.StartTimer(c.Request.Context(), h.userID, problemID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"id":         ts.ID,
		"started_at": ts.StartedAt,
		"problem_id": ts.ProblemID,
	})
}

// StopTimer handles POST /api/timers/stop
func (h *TimersAPI) StopTimer(c *gin.Context) {
	ts, err := h.service.StopTimer(c.Request.Context(), h.userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if ts == nil {
		c.JSON(http.StatusOK, gin.H{"active": false, "elapsed_seconds": 0})
		return
	}

	c.JSON(http.StatusOK, gin.H{"elapsed_seconds": ts.ElapsedSecs})
}

// CurrentTimer handles GET /api/timers/current
func (h *TimersAPI) CurrentTimer(c *gin.Context) {
	ts, err := h.service.CurrentTimer(c.Request.Context(), h.userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if ts == nil {
		c.JSON(http.StatusOK, gin.H{"active": false})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"active":          true,
		"id":              ts.ID,
		"started_at":      ts.StartedAt,
		"elapsed_seconds": postgres.ElapsedNow(ts),
		"problem_id":      ts.ProblemID,
	})
}
