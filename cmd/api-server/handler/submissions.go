package handler

import (
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/tuannm99/judge-loop/internal/domain"
	"github.com/tuannm99/judge-loop/internal/queue"
)

// createSubmissionRequest is the POST /api/submissions request body.
type createSubmissionRequest struct {
	ProblemID string `json:"problem_id" binding:"required"`
	Language  string `json:"language"   binding:"required"`
	Code      string `json:"code"       binding:"required"`
	SessionID string `json:"session_id"` // optional
}

// CreateSubmission handles POST /api/submissions.
// It persists the submission as pending, then enqueues an evaluation job.
func (h *Handler) CreateSubmission(c *gin.Context) {
	var req createSubmissionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	problemID, err := uuid.Parse(req.ProblemID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid problem_id"})
		return
	}

	sub := &domain.Submission{
		UserID:    h.UserID,
		ProblemID: problemID,
		Language:  domain.Language(req.Language),
		Code:      req.Code,
	}
	if req.SessionID != "" {
		if sid, err := uuid.Parse(req.SessionID); err == nil {
			sub.SessionID = &sid
		}
	}

	if err := h.Submissions.Create(c.Request.Context(), sub); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Enqueue evaluation job for the judge-worker.
	task, err := queue.NewEvaluateTask(queue.EvaluatePayload{
		SubmissionID: sub.ID.String(),
		UserID:       sub.UserID.String(),
	})
	if err != nil {
		log.Printf("build evaluate task for %s: %v", sub.ID, err)
	} else if _, err := h.Queue.Enqueue(task); err != nil {
		log.Printf("enqueue evaluate task for %s: %v", sub.ID, err)
	}

	c.JSON(http.StatusCreated, gin.H{
		"submission_id": sub.ID,
		"status":        string(domain.StatusPending),
	})
}

// GetSubmission handles GET /api/submissions/:id
func (h *Handler) GetSubmission(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid submission id"})
		return
	}

	sub, err := h.Submissions.GetByID(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if sub == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "submission not found"})
		return
	}

	c.JSON(http.StatusOK, sub)
}

// ListSubmissions handles GET /api/submissions/history
func (h *Handler) ListSubmissions(c *gin.Context) {
	var problemID *uuid.UUID
	if raw := c.Query("problem_id"); raw != "" {
		if id, err := uuid.Parse(raw); err == nil {
			problemID = &id
		}
	}

	subs, err := h.Submissions.ListByUser(c.Request.Context(), h.UserID, problemID, 20, 0)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"submissions": subs})
}
