package apiserver

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
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
func (h *SubmissionsAPI) CreateSubmission(c *gin.Context) {
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

	var sessionID *uuid.UUID
	if req.SessionID != "" {
		if sid, err := uuid.Parse(req.SessionID); err == nil {
			sessionID = &sid
		}
	}

	sub, err := h.deps.submissions.CreateSubmission(
		c.Request.Context(),
		h.deps.userID,
		problemID,
		req.Language,
		req.Code,
		sessionID,
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"submission_id": sub.ID,
		"status":        string(sub.Status),
	})
}

// GetSubmission handles GET /api/submissions/:id
func (h *SubmissionsAPI) GetSubmission(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid submission id"})
		return
	}

	sub, err := h.deps.submissions.GetSubmission(c.Request.Context(), id)
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
func (h *SubmissionsAPI) ListSubmissions(c *gin.Context) {
	var problemID *uuid.UUID
	if raw := c.Query("problem_id"); raw != "" {
		if id, err := uuid.Parse(raw); err == nil {
			problemID = &id
		}
	}

	subs, err := h.deps.submissions.ListSubmissions(c.Request.Context(), h.deps.userID, problemID, 20, 0)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"submissions": subs})
}
