package application

import (
	"context"
	"log"
	"time"

	"github.com/google/uuid"
	"github.com/tuannm99/judge-loop/internal/domain"
	inport "github.com/tuannm99/judge-loop/internal/port/in"
	outport "github.com/tuannm99/judge-loop/internal/port/out"
)

type SubmissionService struct {
	submissions outport.SubmissionRepository
	publisher   outport.EvaluationPublisher
	evaluator   inport.EvaluationService
	timeLimit   int
}

var _ inport.SubmissionService = (*SubmissionService)(nil)

func NewSubmissionService(
	submissions outport.SubmissionRepository,
	publisher outport.EvaluationPublisher,
	evaluator inport.EvaluationService,
	timeLimit int,
) *SubmissionService {
	return &SubmissionService{
		submissions: submissions,
		publisher:   publisher,
		evaluator:   evaluator,
		timeLimit:   timeLimit,
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
		s.scheduleFallback(sub.ID, sub.UserID)
		return sub, nil
	}
	if err := s.publisher.PublishEvaluation(outport.EvaluateSubmissionJob{
		SubmissionID: sub.ID.String(),
		UserID:       sub.UserID.String(),
	}); err != nil {
		log.Printf("submission %s: queue publish failed, falling back to local evaluation: %v", sub.ID, err)
		s.scheduleFallback(sub.ID, sub.UserID)
		return sub, nil
	}
	s.scheduleFallback(sub.ID, sub.UserID)
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

func (s *SubmissionService) scheduleFallback(submissionID, userID uuid.UUID) {
	if s.evaluator == nil || s.timeLimit <= 0 {
		return
	}

	go func() {
		time.Sleep(2 * time.Second)
		if err := s.evaluator.EvaluateSubmission(context.Background(), submissionID, userID, s.timeLimit); err != nil {
			log.Printf("submission %s: fallback evaluation failed: %v", submissionID, err)
		}
	}()
}
