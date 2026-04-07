package out

import (
	"context"

	"github.com/google/uuid"
)

type ReviewRepository interface {
	GetDue(ctx context.Context, userID uuid.UUID) ([]DueReview, error)
	Upsert(ctx context.Context, userID, problemID uuid.UUID) error
	// Reset sets next_review_at to tomorrow and interval_days to 1 when the user
	// regresses on a problem. It is a no-op if no review schedule exists yet.
	Reset(ctx context.Context, userID, problemID uuid.UUID) error
}
