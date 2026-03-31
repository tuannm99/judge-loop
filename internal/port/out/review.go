package out

import (
	"context"

	"github.com/google/uuid"
)

type ReviewRepository interface {
	GetDue(ctx context.Context, userID uuid.UUID) ([]DueReview, error)
	Upsert(ctx context.Context, userID, problemID uuid.UUID) error
}
