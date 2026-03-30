package apiserver

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/tuannm99/judge-loop/internal/domain"
	outport "github.com/tuannm99/judge-loop/internal/port/out"
)

// ListProblems handles GET /api/problems
func (h *ProblemsAPI) ListProblems(c *gin.Context) {
	f := outport.ProblemFilter{
		Tag:     c.Query("tag"),
		Pattern: c.Query("pattern"),
	}

	if d := c.Query("difficulty"); d != "" {
		diff := domain.Difficulty(d)
		f.Difficulty = &diff
	}
	if p := c.Query("provider"); p != "" {
		prov := domain.Provider(p)
		f.Provider = &prov
	}
	if l := c.Query("limit"); l != "" {
		n, err := strconv.Atoi(l)
		if err == nil && n > 0 {
			f.Limit = n
		}
	}
	if o := c.Query("offset"); o != "" {
		n, err := strconv.Atoi(o)
		if err == nil && n >= 0 {
			f.Offset = n
		}
	}

	problems, total, err := h.deps.problems.ListProblems(c.Request.Context(), f)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"problems": problems,
		"total":    total,
	})
}

// GetProblem handles GET /api/problems/:id
// Tries UUID parse first; falls back to slug lookup.
func (h *ProblemsAPI) GetProblem(c *gin.Context) {
	problem, err := h.deps.problems.GetProblem(c.Request.Context(), c.Param("id"))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if problem == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "problem not found"})
		return
	}

	c.JSON(http.StatusOK, problem)
}

// SuggestProblem handles GET /api/problems/suggest
// Returns a random unsolved problem, preferring weak patterns from the user profile.
// In Milestone 2, no profile is consulted — patterns list is empty.
func (h *ProblemsAPI) SuggestProblem(c *gin.Context) {
	problem, err := h.deps.problems.SuggestProblem(c.Request.Context(), h.deps.userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if problem == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "no unsolved problems available"})
		return
	}
	c.JSON(http.StatusOK, problem)
}
