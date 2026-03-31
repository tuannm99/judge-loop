package localagent

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// NewRouter creates the local-agent Gin engine with all routes registered.
func NewRouter(handler *Handler) *gin.Engine {
	router := gin.New()
	router.Use(gin.Logger(), gin.Recovery())
	RegisterRoutes(router, handler)
	return router
}

// RegisterRoutes registers all local-agent routes on the given router.
func RegisterRoutes(router gin.IRouter, handler *Handler) {
	router.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	local := router.Group("/local")
	local.GET("/status/today", handler.StatusToday)
	local.GET("/timer/current", handler.CurrentTimer)
	local.POST("/timer/start", handler.StartTimer)
	local.POST("/timer/stop", handler.StopTimer)
	local.POST("/submit", handler.Submit)
	local.POST("/sync", handler.Sync)
	local.GET("/submissions/:id", handler.GetSubmissionStatus)
}
