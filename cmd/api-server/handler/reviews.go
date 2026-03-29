package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// GetReviewsToday handles GET /api/reviews/today
func (h *Handler) GetReviewsToday(c *gin.Context) {
	reviews, err := h.Reviews.GetDue(c.Request.Context(), h.UserID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"reviews": reviews})
}
