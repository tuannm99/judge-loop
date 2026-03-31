package in

import (
	"context"

	"github.com/google/uuid"
	"github.com/tuannm99/judge-loop/internal/domain"
)

type SubmissionService interface {
	CreateSubmission(
		ctx context.Context,
		userID uuid.UUID,
		problemID uuid.UUID,
		language, code string,
		sessionID *uuid.UUID,
	) (*domain.Submission, error)
	GetSubmission(ctx context.Context, id uuid.UUID) (*domain.Submission, error)
	ListSubmissions(
		ctx context.Context,
		userID uuid.UUID,
		problemID *uuid.UUID,
		limit, offset int,
	) ([]domain.Submission, error)
}
