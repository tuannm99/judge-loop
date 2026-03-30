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

	apiHandlers := apiserver.New(db, cfg.UserID, queueClient)

	s := &Server{cfg: cfg, db: db, router: router}
	s.registerRoutes(apiHandlers)
	return s
}

// Run starts the HTTP server.
func (s *Server) Run() error {
	addr := fmt.Sprintf(":%s", s.cfg.Port)
	return s.router.Run(addr)
}

func (s *Server) registerRoutes(apiHandlers *apiserver.API) {
	s.router.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	api := s.router.Group("/api")
	{
		// problems — register /suggest before /:id so the static path wins
		problems := api.Group("/problems")
		problems.GET("", apiHandlers.Problems.ListProblems)
		problems.GET("/suggest", apiHandlers.Problems.SuggestProblem)
		problems.GET("/:id", apiHandlers.Problems.GetProblem)

		// submissions — register /history before /:id
		subs := api.Group("/submissions")
		subs.POST("", apiHandlers.Submissions.CreateSubmission)
		subs.GET("/history", apiHandlers.Submissions.ListSubmissions)
		subs.GET("/:id", apiHandlers.Submissions.GetSubmission)

		// progress & streak
		api.GET("/progress/today", apiHandlers.Progress.GetProgressToday)
		api.GET("/streak", apiHandlers.Progress.GetStreak)

		// timers
		timers := api.Group("/timers")
		timers.POST("/start", apiHandlers.Timers.StartTimer)
		timers.POST("/stop", apiHandlers.Timers.StopTimer)
		timers.GET("/current", apiHandlers.Timers.CurrentTimer)

		// reviews
		api.GET("/reviews/today", apiHandlers.Reviews.GetReviewsToday)

		// registry
		api.POST("/registry/sync", apiHandlers.Registry.SyncRegistry)
		api.GET("/registry/version", apiHandlers.Registry.GetRegistryVersion)
	}
}
