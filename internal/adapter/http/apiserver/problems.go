package apiserver

import (
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/tuannm99/judge-loop/internal/domain"
	outport "github.com/tuannm99/judge-loop/internal/port/out"
)

type contributeProblemRequest struct {
	Provider            domain.Provider             `json:"provider"       binding:"required"`
	ExternalID          string                      `json:"external_id"    binding:"required"`
	Slug                string                      `json:"slug"           binding:"required"`
	Title               string                      `json:"title"          binding:"required"`
	Difficulty          domain.Difficulty           `json:"difficulty"     binding:"required"`
	Tags                []string                    `json:"tags"`
	LegacyPatternTags   []string                    `json:"pattern_tags"`
	SourceURL           string                      `json:"source_url"     binding:"required"`
	EstimatedTime       int                         `json:"estimated_time"`
	DescriptionMarkdown string                      `json:"description_markdown"`
	StarterCode         map[string]string           `json:"starter_code"`
	Version             int                         `json:"version"`
	TestCases           []contributeTestCaseRequest `json:"test_cases"     binding:"required,min=1"`
}

type updateProblemRequest struct {
	Provider            domain.Provider             `json:"provider"       binding:"required"`
	ExternalID          string                      `json:"external_id"    binding:"required"`
	Slug                string                      `json:"slug"           binding:"required"`
	Title               string                      `json:"title"          binding:"required"`
	Difficulty          domain.Difficulty           `json:"difficulty"     binding:"required"`
	Tags                []string                    `json:"tags"`
	LegacyPatternTags   []string                    `json:"pattern_tags"`
	SourceURL           string                      `json:"source_url"     binding:"required"`
	EstimatedTime       int                         `json:"estimated_time"`
	DescriptionMarkdown string                      `json:"description_markdown"`
	StarterCode         map[string]string           `json:"starter_code"`
	TestCases           []contributeTestCaseRequest `json:"test_cases"`
}

type contributeTestCaseRequest struct {
	Input    string `json:"input"     binding:"required"`
	Expected string `json:"expected"  binding:"required"`
	IsHidden bool   `json:"is_hidden"`
}

type problemLabelRequest struct {
	Kind string `json:"kind"`
	Slug string `json:"slug" binding:"required"`
	Name string `json:"name"`
}

type updateProblemLabelRequest struct {
	Slug string `json:"slug" binding:"required"`
	Name string `json:"name"`
}

// ListProblemLabels handles GET /api/problems/labels.
func (h *ProblemsAPI) ListProblemLabels(c *gin.Context) {
	tags, err := h.service.ListProblemLabels(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"tags": tags,
	})
}

// ListProblemLabelRecords handles GET /api/problem-labels.
func (h *ProblemsAPI) ListProblemLabelRecords(c *gin.Context) {
	kind := normalizeProblemLabelKind(c.Query("kind"))
	if kind == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "kind must be tag or pattern"})
		return
	}

	labels, err := h.service.ListProblemLabelRecords(c.Request.Context(), kind)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"labels": labels})
}

// CreateProblemLabel handles POST /api/problem-labels.
func (h *ProblemsAPI) CreateProblemLabel(c *gin.Context) {
	var req problemLabelRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	kind := normalizeProblemLabelKind(req.Kind)
	if kind == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "kind must be tag or pattern"})
		return
	}

	label, err := h.service.CreateProblemLabel(c.Request.Context(), kind, req.Slug, req.Name)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, label)
}

// UpdateProblemLabel handles PUT /api/problem-labels/:id.
func (h *ProblemsAPI) UpdateProblemLabel(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid label id"})
		return
	}

	var req updateProblemLabelRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	label, err := h.service.UpdateProblemLabel(c.Request.Context(), id, req.Slug, req.Name)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, label)
}

// DeleteProblemLabel handles DELETE /api/problem-labels/:id.
func (h *ProblemsAPI) DeleteProblemLabel(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid label id"})
		return
	}

	if err := h.service.DeleteProblemLabel(c.Request.Context(), id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.AbortWithStatus(http.StatusNoContent)
}

// ListProblems handles GET /api/problems
func (h *ProblemsAPI) ListProblems(c *gin.Context) {
	tags := append(c.QueryArray("tag"), c.QueryArray("tags")...)
	tags = append(tags, c.QueryArray("pattern")...)
	tags = append(tags, c.QueryArray("patterns")...)

	f := outport.ProblemFilter{
		Tags: tags,
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

	problems, total, err := h.service.ListProblems(c.Request.Context(), f)
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
	problem, err := h.service.GetProblem(c.Request.Context(), c.Param("id"))
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

// GetProblemTestCases handles GET /api/problems/:id/test-cases
func (h *ProblemsAPI) GetProblemTestCases(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid problem id"})
		return
	}

	testCases, err := h.service.GetProblemTestCases(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"test_cases": testCases})
}

// UpdateProblem handles PUT /api/problems/:id
func (h *ProblemsAPI) UpdateProblem(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid problem id"})
		return
	}

	var req updateProblemRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	manifest := domain.ProblemManifest{
		Provider:            req.Provider,
		ExternalID:          req.ExternalID,
		Slug:                req.Slug,
		Title:               req.Title,
		Difficulty:          req.Difficulty,
		Tags:                mergeProblemTags(req.Tags, req.LegacyPatternTags),
		SourceURL:           req.SourceURL,
		EstimatedTime:       req.EstimatedTime,
		DescriptionMarkdown: req.DescriptionMarkdown,
		StarterCode:         req.StarterCode,
	}
	testCases := make([]domain.TestCase, 0, len(req.TestCases))
	for i, tc := range req.TestCases {
		testCases = append(testCases, domain.TestCase{
			Input:    tc.Input,
			Expected: tc.Expected,
			IsHidden: tc.IsHidden,
			OrderIdx: i,
		})
	}

	var problem *domain.Problem
	if req.TestCases != nil {
		problem, err = h.service.UpdateProblemWithTestCases(c.Request.Context(), id, manifest, testCases)
	} else {
		problem, err = h.service.UpdateProblem(c.Request.Context(), id, manifest)
	}
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

// SuggestProblem handles GET /api/problems/suggest.
// Returns a random unsolved problem. Later personalization can use weak tags to bias selection.
func (h *ProblemsAPI) SuggestProblem(c *gin.Context) {
	problem, err := h.service.SuggestProblem(c.Request.Context(), h.userID)
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

// ContributeProblem handles POST /api/problems/contribute
func (h *ProblemsAPI) ContributeProblem(c *gin.Context) {
	var req contributeProblemRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	testCases := make([]domain.TestCase, 0, len(req.TestCases))
	for i, tc := range req.TestCases {
		testCases = append(testCases, domain.TestCase{
			Input:    tc.Input,
			Expected: tc.Expected,
			IsHidden: tc.IsHidden,
			OrderIdx: i,
		})
	}

	problem, err := h.service.ContributeProblem(c.Request.Context(), domain.ProblemManifest{
		Provider:            req.Provider,
		ExternalID:          req.ExternalID,
		Slug:                req.Slug,
		Title:               req.Title,
		Difficulty:          req.Difficulty,
		Tags:                mergeProblemTags(req.Tags, req.LegacyPatternTags),
		SourceURL:           req.SourceURL,
		EstimatedTime:       req.EstimatedTime,
		DescriptionMarkdown: req.DescriptionMarkdown,
		StarterCode:         req.StarterCode,
		Version:             req.Version,
	}, testCases)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if problem == nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "problem not found after contribution"})
		return
	}

	c.JSON(http.StatusCreated, problem)
}

func normalizeProblemLabelKind(kind string) string {
	switch strings.TrimSpace(kind) {
	case "", "tag", "pattern":
		return "tag"
	default:
		return ""
	}
}

func mergeProblemTags(groups ...[]string) []string {
	seen := make(map[string]struct{})
	out := make([]string, 0)
	for _, group := range groups {
		for _, value := range group {
			value = strings.TrimSpace(value)
			if value == "" {
				continue
			}
			if _, ok := seen[value]; ok {
				continue
			}
			seen[value] = struct{}{}
			out = append(out, value)
		}
	}
	return out
}
