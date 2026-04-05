package apiserver

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// NewRouter creates the api-server Gin engine with all routes registered.
func NewRouter(api *API) *gin.Engine {
	router := gin.New()
	router.Use(gin.Logger(), gin.Recovery())
	RegisterRoutes(router, api)
	return router
}

// RegisterRoutes registers all api-server routes on the given router.
func RegisterRoutes(router gin.IRouter, api *API) {
	router.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	v1 := router.Group("/api")
	{
		problems := v1.Group("/problems")
		problems.GET("", api.Problems.ListProblems)
		problems.GET("/labels", api.Problems.ListProblemLabels)
		problems.POST("/contribute", api.Problems.ContributeProblem)
		problems.GET("/suggest", api.Problems.SuggestProblem)
		problems.GET("/:id", api.Problems.GetProblem)

		problemLabels := v1.Group("/problem-labels")
		problemLabels.GET("", api.Problems.ListProblemLabelRecords)
		problemLabels.POST("", api.Problems.CreateProblemLabel)
		problemLabels.PUT("/:id", api.Problems.UpdateProblemLabel)
		problemLabels.DELETE("/:id", api.Problems.DeleteProblemLabel)

		submissions := v1.Group("/submissions")
		submissions.POST("", api.Submissions.CreateSubmission)
		submissions.GET("/history", api.Submissions.ListSubmissions)
		submissions.GET("/:id", api.Submissions.GetSubmission)

		v1.GET("/progress/today", api.Progress.GetProgressToday)
		v1.GET("/streak", api.Progress.GetStreak)

		timers := v1.Group("/timers")
		timers.POST("/start", api.Timers.StartTimer)
		timers.POST("/stop", api.Timers.StopTimer)
		timers.GET("/current", api.Timers.CurrentTimer)

		v1.GET("/reviews/today", api.Reviews.GetReviewsToday)
		v1.POST("/registry/sync", api.Registry.SyncRegistry)
		v1.GET("/registry/version", api.Registry.GetRegistryVersion)
	}
}
