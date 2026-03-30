package main

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
)

// Server wires together the Gin router and all dependencies.
type Server struct {
	cfg     Config
	router  *gin.Engine
	handler *Handler
}

// NewServer initialises the server and registers all routes.
func NewServer(cfg Config) *Server {
	client := NewAPIClient(cfg.ServerURL, cfg.UserID)
	h := NewHandler(client, cfg)

	router := gin.New()
	router.Use(gin.Logger(), gin.Recovery())

	s := &Server{cfg: cfg, router: router, handler: h}
	s.registerRoutes()
	return s
}

// Run starts the HTTP server bound to localhost only.
func (s *Server) Run() error {
	addr := fmt.Sprintf("127.0.0.1:%d", s.cfg.Port)
	return s.router.Run(addr)
}

func (s *Server) registerRoutes() {
	h := s.handler

	s.router.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	local := s.router.Group("/local")
	local.GET("/status/today", h.StatusToday)
	local.GET("/timer/current", h.CurrentTimer)
	local.POST("/timer/start", h.StartTimer)
	local.POST("/timer/stop", h.StopTimer)
	local.POST("/submit", h.Submit)
	local.POST("/sync", h.Sync)
	local.GET("/submissions/:id", h.GetSubmissionStatus)
}
