package handler

import (
	"context"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/tuannm99/judge-loop/internal/domain"
)

// createSubmissionRequest is the POST /api/submissions request body.
type createSubmissionRequest struct {
	ProblemID string `json:"problem_id" binding:"required"`
	Language  string `json:"language"   binding:"required"`
	Code      string `json:"code"       binding:"required"`
	SessionID string `json:"session_id"` // optional
}

// CreateSubmission handles POST /api/submissions
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

	// evaluate asynchronously with a mock evaluator (Milestone 6 replaces this)
	go func() {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		h.evaluateMock(ctx, *sub)
	}()

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

// evaluateMock simulates verdict evaluation.
// It accepts any non-empty code submission. This will be replaced in Milestone 6
// by the real judge-worker running code in a Docker sandbox.
func (h *Handler) evaluateMock(ctx context.Context, sub domain.Submission) {
	time.Sleep(500 * time.Millisecond)

	status := string(domain.StatusAccepted)
	verdict := string(domain.VerdictAccepted)
	passed := 5
	total := 5

	if strings.TrimSpace(sub.Code) == "" {
		status = string(domain.StatusWrongAnswer)
		verdict = string(domain.VerdictWrongAnswer)
		passed = 0
	}

	now := time.Now()
	_ = h.Submissions.UpdateVerdict(ctx, sub.ID, status, verdict, passed, total, 42, "", &now)
	_ = h.Sessions.RecordSubmission(ctx, sub.UserID, status == string(domain.StatusAccepted))
}
