package apiserver

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// GetReviewsToday handles GET /api/reviews/today
func (h *ReviewsAPI) GetReviewsToday(c *gin.Context) {
	reviews, err := h.deps.reviews.GetReviewsToday(c.Request.Context(), h.deps.userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"reviews": reviews})
}
