package main

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/hibiken/asynq"
	apiserver "github.com/tuannm99/judge-loop/internal/adapter/http/apiserver"
	postgres "github.com/tuannm99/judge-loop/internal/infrastructure/postgres"
)

// Server wires together the Gin router and all dependencies.
type Server struct {
	cfg    Config
	db     *postgres.DB
	router *gin.Engine
}

// NewServer initialises the server with a database connection and registers all routes.
func NewServer(cfg Config, db *postgres.DB, queueClient *asynq.Client) *Server {
	router := gin.New()
	router.Use(gin.Logger(), gin.Recovery())

	h := apiserver.New(db, cfg.UserID, queueClient)

	s := &Server{cfg: cfg, db: db, router: router}
	s.registerRoutes(h)
	return s
}

// Run starts the HTTP server.
func (s *Server) Run() error {
	addr := fmt.Sprintf(":%s", s.cfg.Port)
	return s.router.Run(addr)
}

func (s *Server) registerRoutes(h *apiserver.Handler) {
	s.router.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	api := s.router.Group("/api")
	{
		// problems — register /suggest before /:id so the static path wins
		problems := api.Group("/problems")
		problems.GET("", h.ListProblems)
		problems.GET("/suggest", h.SuggestProblem)
		problems.GET("/:id", h.GetProblem)

		// submissions — register /history before /:id
		subs := api.Group("/submissions")
		subs.POST("", h.CreateSubmission)
		subs.GET("/history", h.ListSubmissions)
		subs.GET("/:id", h.GetSubmission)

		// progress & streak
		api.GET("/progress/today", h.GetProgressToday)
		api.GET("/streak", h.GetStreak)

		// timers
		timers := api.Group("/timers")
		timers.POST("/start", h.StartTimer)
		timers.POST("/stop", h.StopTimer)
		timers.GET("/current", h.CurrentTimer)

		// reviews
		api.GET("/reviews/today", h.GetReviewsToday)

		// registry
		api.POST("/registry/sync", h.SyncRegistry)
		api.GET("/registry/version", h.GetRegistryVersion)
	}
}
