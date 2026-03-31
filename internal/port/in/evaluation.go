package in

import (
	"context"

	"github.com/google/uuid"
)

type EvaluationService interface {
	EvaluateSubmission(ctx context.Context, submissionID, userID uuid.UUID, timeLimitSecs int) error
}
