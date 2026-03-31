package apiserver

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/tuannm99/judge-loop/internal/domain"
)

// registrySyncRequest is the body for POST /api/registry/sync.
type registrySyncRequest struct {
	Version   string                   `json:"version"    binding:"required"`
	UpdatedAt time.Time                `json:"updated_at"`
	Problems  []domain.ProblemManifest `json:"problems"   binding:"required"`
	Manifests []domain.ManifestRef     `json:"manifests"`
}

// SyncRegistry handles POST /api/registry/sync.
// It upserts all problems from the supplied manifests and records the registry version.
func (h *RegistryAPI) SyncRegistry(c *gin.Context) {
	var req registrySyncRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	synced, err := h.service.SyncRegistry(
		c.Request.Context(),
		req.Version,
		req.UpdatedAt,
		req.Problems,
		req.Manifests,
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"version": req.Version,
		"synced":  synced,
	})
}

// GetRegistryVersion handles GET /api/registry/version.
func (h *RegistryAPI) GetRegistryVersion(c *gin.Context) {
	row, err := h.service.GetRegistryVersion(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if row == nil {
		c.JSON(http.StatusOK, gin.H{"version": "none", "synced_at": nil})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"version":   row.Version,
		"synced_at": row.SyncedAt,
	})
}
