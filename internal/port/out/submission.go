package out

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/tuannm99/judge-loop/internal/domain"
)

type SubmissionRepository interface {
	Create(ctx context.Context, sub *domain.Submission) error
	GetByID(ctx context.Context, id uuid.UUID) (*domain.Submission, error)
	TryStartEvaluation(ctx context.Context, id uuid.UUID) (bool, error)
	UpdateVerdict(
		ctx context.Context,
		id uuid.UUID,
		status, verdict string,
		passed, total int,
		runtimeMS int64,
		errMsg string,
		evaluatedAt *time.Time,
	) error
	ListByUser(
		ctx context.Context,
		userID uuid.UUID,
		problemID *uuid.UUID,
		limit, offset int,
	) ([]domain.Submission, error)
}
