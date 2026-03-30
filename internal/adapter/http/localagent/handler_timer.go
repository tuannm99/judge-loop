package localagent

import (
	"context"
	"log"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type startTimerRequest struct {
	ProblemID string `json:"problem_id"`
}

// StartTimer handles POST /local/timer/start.
// Starts the local in-memory timer immediately; notifies api-server best-effort.
func (h *Handler) StartTimer(c *gin.Context) {
	var req startTimerRequest
	_ = c.ShouldBindJSON(&req) // body is optional

	var problemID *uuid.UUID
	if req.ProblemID != "" {
		if id, err := uuid.Parse(req.ProblemID); err == nil {
			problemID = &id
		}
	}

	entry := h.timer.Start(problemID)

	// notify api-server best-effort — never block the local response
	go func() {
		ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
		defer cancel()
		if err := h.client.StartTimer(ctx, req.ProblemID); err != nil {
			log.Printf("timer: sync start to server failed: %v", err)
		}
	}()

	c.JSON(http.StatusOK, gin.H{
		"id":         entry.ID,
		"started_at": entry.StartedAt,
		"problem_id": entry.ProblemID,
	})
}

// StopTimer handles POST /local/timer/stop.
// Stops the local timer; notifies api-server best-effort.
func (h *Handler) StopTimer(c *gin.Context) {
	entry, ok := h.timer.Stop()
	if !ok {
		c.JSON(http.StatusOK, gin.H{"active": false, "elapsed_seconds": 0})
		return
	}

	elapsed := int(time.Since(entry.StartedAt).Seconds())

	go func() {
		ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
		defer cancel()
		if err := h.client.StopTimer(ctx); err != nil {
			log.Printf("timer: sync stop to server failed: %v", err)
		}
	}()

	c.JSON(http.StatusOK, gin.H{"elapsed_seconds": elapsed})
}

// CurrentTimer handles GET /local/timer/current.
// Returns local state if a timer is running; falls back to api-server.
func (h *Handler) CurrentTimer(c *gin.Context) {
	if active := h.timer.Active(); active != nil {
		c.JSON(http.StatusOK, gin.H{
			"active":          true,
			"id":              active.ID,
			"started_at":      active.StartedAt,
			"elapsed_seconds": h.timer.ElapsedSecs(),
			"problem_id":      active.ProblemID,
		})
		return
	}

	// no local timer — ask the server
	ctx, cancel := context.WithTimeout(c.Request.Context(), 3*time.Second)
	defer cancel()

	resp, err := h.client.CurrentTimer(ctx)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"active": false})
		return
	}

	c.JSON(http.StatusOK, resp)
}
