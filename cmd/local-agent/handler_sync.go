package main

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/tuannm99/judge-loop/internal/domain"
	"github.com/tuannm99/judge-loop/internal/registry"
)

// Sync handles POST /local/sync.
// It reads the local registry manifests and upserts them into the api-server.
func (h *Handler) Sync(c *gin.Context) {
	if h.registryPath == "" {
		c.JSON(http.StatusServiceUnavailable, gin.H{
			"synced":  false,
			"message": "registry_path not configured",
		})
		return
	}

	ctx, cancel := context.WithTimeout(c.Request.Context(), 30*time.Second)
	defer cancel()

	// 1. Load index.
	idx, err := registry.LoadIndex(h.registryPath)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"synced":  false,
			"message": fmt.Sprintf("load registry index: %v", err),
		})
		return
	}

	// 2. Load all problems from provider manifests.
	problems, err := registry.LoadAllProblems(h.registryPath, idx)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"synced":  false,
			"message": fmt.Sprintf("load registry manifests: %v", err),
		})
		return
	}

	// 3. Forward to api-server for upsert.
	resp, err := h.client.RegistrySync(ctx, RegistrySyncRequest{
		Version:   idx.Version,
		Problems:  problems,
		Manifests: toRefDTOs(idx.Manifests),
	})
	if err != nil {
		c.JSON(http.StatusBadGateway, gin.H{
			"synced":  false,
			"message": fmt.Sprintf("api-server registry sync: %v", err),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"synced":   true,
		"version":  resp.Version,
		"problems": resp.Synced,
		"message":  fmt.Sprintf("Registry synced: %d problems (version %s)", resp.Synced, resp.Version),
	})
}

// toRefDTOs converts domain.ManifestRef slice to the same type (they're identical now).
func toRefDTOs(refs []domain.ManifestRef) []domain.ManifestRef {
	return refs
}
