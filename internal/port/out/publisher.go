package out

import (
	"context"

	"github.com/google/uuid"
)

type EvaluationPublisher interface {
	PublishEvaluation(job EvaluateSubmissionJob) error
}

type EvaluationJobQueue interface {
	ClaimEvaluationJob(ctx context.Context, workerID string) (*EvaluationJob, error)
	CompleteEvaluationJob(ctx context.Context, id uuid.UUID) error
	FailEvaluationJob(ctx context.Context, id uuid.UUID, errMsg string) error
}
