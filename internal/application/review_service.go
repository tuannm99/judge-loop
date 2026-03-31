package application

import (
	"context"

	"github.com/google/uuid"
	inport "github.com/tuannm99/judge-loop/internal/port/in"
	outport "github.com/tuannm99/judge-loop/internal/port/out"
)

type ReviewService struct {
	reviews outport.ReviewRepository
}

var _ inport.ReviewService = (*ReviewService)(nil)

func NewReviewService(reviews outport.ReviewRepository) *ReviewService {
	return &ReviewService{reviews: reviews}
}

func (s *ReviewService) GetReviewsToday(ctx context.Context, userID uuid.UUID) ([]outport.DueReview, error) {
	return s.reviews.GetDue(ctx, userID)
}
