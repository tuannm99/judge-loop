package in

import (
	"context"

	"github.com/google/uuid"
	out "github.com/tuannm99/judge-loop/internal/port/out"
)

type ReviewService interface {
	GetReviewsToday(ctx context.Context, userID uuid.UUID) ([]out.DueReview, error)
}
