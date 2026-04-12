package application

import (
	"context"
	"fmt"
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
	normalizedLanguage := domain.NormalizeSubmissionLanguage(language)
	if !domain.IsSupportedSubmissionLanguage(string(normalizedLanguage)) {
		return nil, fmt.Errorf("%w: %s", domain.ErrUnsupportedSubmissionLanguage, language)
	}

	sub := &domain.Submission{
		UserID:    userID,
		ProblemID: problemID,
		SessionID: sessionID,
		Language:  normalizedLanguage,
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
	timeLimit := s.timeLimit
	go func() {
		time.Sleep(2 * time.Second)
		// Bound context to time limit + 30s buffer so the goroutine cannot run forever.
		ctx, cancel := context.WithTimeout(context.Background(), time.Duration(timeLimit+30)*time.Second)
		defer cancel()
		if err := s.evaluator.EvaluateSubmission(ctx, submissionID, userID, timeLimit); err != nil {
			log.Printf("submission %s: fallback evaluation failed: %v", submissionID, err)
		}
	}()
}
