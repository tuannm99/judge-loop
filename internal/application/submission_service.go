package application

import (
	"context"

	"github.com/google/uuid"
	"github.com/tuannm99/judge-loop/internal/domain"
	inport "github.com/tuannm99/judge-loop/internal/port/in"
	outport "github.com/tuannm99/judge-loop/internal/port/out"
)

type SubmissionService struct {
	submissions outport.SubmissionRepository
	publisher   outport.EvaluationPublisher
}

var _ inport.SubmissionService = (*SubmissionService)(nil)

func NewSubmissionService(
	submissions outport.SubmissionRepository,
	publisher outport.EvaluationPublisher,
) *SubmissionService {
	return &SubmissionService{
		submissions: submissions,
		publisher:   publisher,
	}
}

func (s *SubmissionService) CreateSubmission(
	ctx context.Context,
	userID uuid.UUID,
	problemID uuid.UUID,
	language, code string,
	sessionID *uuid.UUID,
) (*domain.Submission, error) {
	sub := &domain.Submission{
		UserID:    userID,
		ProblemID: problemID,
		SessionID: sessionID,
		Language:  domain.Language(language),
		Code:      code,
	}
	if err := s.submissions.Create(ctx, sub); err != nil {
		return nil, err
	}
	if s.publisher == nil {
		return sub, nil
	}
	_ = s.publisher.PublishEvaluation(outport.EvaluateSubmissionJob{
		SubmissionID: sub.ID.String(),
		UserID:       sub.UserID.String(),
	})
	return sub, nil
}

func (s *SubmissionService) GetSubmission(ctx context.Context, id uuid.UUID) (*domain.Submission, error) {
	return s.submissions.GetByID(ctx, id)
}

func (s *SubmissionService) ListSubmissions(
	ctx context.Context,
	userID uuid.UUID,
	problemID *uuid.UUID,
	limit, offset int,
) ([]domain.Submission, error) {
	return s.submissions.ListByUser(ctx, userID, problemID, limit, offset)
}
