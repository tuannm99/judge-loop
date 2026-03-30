package application

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/tuannm99/judge-loop/internal/domain"
	inport "github.com/tuannm99/judge-loop/internal/port/in"
	outport "github.com/tuannm99/judge-loop/internal/port/out"
)

type APIService struct {
	problems    outport.ProblemRepository
	submissions outport.SubmissionRepository
	sessions    outport.SessionRepository
	reviews     outport.ReviewRepository
	registry    outport.RegistryRepository
	publisher   outport.EvaluationPublisher
}

func NewAPIService(
	problems outport.ProblemRepository,
	submissions outport.SubmissionRepository,
	sessions outport.SessionRepository,
	reviews outport.ReviewRepository,
	registry outport.RegistryRepository,
	publisher outport.EvaluationPublisher,
) *APIService {
	return &APIService{
		problems:    problems,
		submissions: submissions,
		sessions:    sessions,
		reviews:     reviews,
		registry:    registry,
		publisher:   publisher,
	}
}

func (s *APIService) ListProblems(ctx context.Context, filter outport.ProblemFilter) ([]domain.Problem, int, error) {
	return s.problems.List(ctx, filter)
}

func (s *APIService) GetProblem(ctx context.Context, rawID string) (*domain.Problem, error) {
	if id, err := uuid.Parse(rawID); err == nil {
		return s.problems.GetByID(ctx, id)
	}
	return s.problems.GetBySlug(ctx, rawID)
}

func (s *APIService) SuggestProblem(ctx context.Context, userID uuid.UUID) (*domain.Problem, error) {
	return s.problems.Suggest(ctx, userID, nil)
}

func (s *APIService) CreateSubmission(
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
	if err := s.publisher.PublishEvaluation(outport.EvaluateSubmissionJob{
		SubmissionID: sub.ID.String(),
		UserID:       sub.UserID.String(),
	}); err != nil {
		return nil, fmt.Errorf("publish evaluation job: %w", err)
	}
	return sub, nil
}

func (s *APIService) GetSubmission(ctx context.Context, id uuid.UUID) (*domain.Submission, error) {
	return s.submissions.GetByID(ctx, id)
}

func (s *APIService) ListSubmissions(
	ctx context.Context,
	userID uuid.UUID,
	problemID *uuid.UUID,
	limit, offset int,
) ([]domain.Submission, error) {
	return s.submissions.ListByUser(ctx, userID, problemID, limit, offset)
}

func (s *APIService) GetProgressToday(ctx context.Context, userID uuid.UUID) (inport.ProgressToday, error) {
	ds, err := s.sessions.GetOrCreateToday(ctx, userID)
	if err != nil {
		return inport.ProgressToday{}, err
	}
	streak, err := s.sessions.GetStreak(ctx, userID)
	if err != nil {
		return inport.ProgressToday{}, err
	}
	return inport.ProgressToday{
		Date:             ds.Date.Format("2006-01-02"),
		Solved:           ds.SolvedCount,
		Attempted:        ds.AttemptedCount,
		TimeSpentMinutes: ds.TimeSpentSecs / 60,
		Streak:           streak.Current,
	}, nil
}

func (s *APIService) GetStreak(ctx context.Context, userID uuid.UUID) (outport.StreakInfo, error) {
	return s.sessions.GetStreak(ctx, userID)
}

func (s *APIService) StartTimer(
	ctx context.Context,
	userID uuid.UUID,
	problemID *uuid.UUID,
) (*domain.TimerSession, error) {
	return s.sessions.StartTimer(ctx, userID, problemID)
}

func (s *APIService) StopTimer(ctx context.Context, userID uuid.UUID) (*domain.TimerSession, error) {
	return s.sessions.StopTimer(ctx, userID)
}

func (s *APIService) CurrentTimer(ctx context.Context, userID uuid.UUID) (*domain.TimerSession, error) {
	return s.sessions.ActiveTimer(ctx, userID)
}

func (s *APIService) GetReviewsToday(ctx context.Context, userID uuid.UUID) ([]outport.DueReview, error) {
	return s.reviews.GetDue(ctx, userID)
}

func (s *APIService) SyncRegistry(
	ctx context.Context,
	version string,
	updatedAt time.Time,
	problems []domain.ProblemManifest,
	manifests []domain.ManifestRef,
) (int, error) {
	synced := 0
	for _, m := range problems {
		if err := s.problems.UpsertFromManifest(ctx, m); err != nil {
			return 0, err
		}
		synced++
	}
	if updatedAt.IsZero() {
		updatedAt = time.Now()
	}
	if err := s.registry.Save(ctx, version, updatedAt, manifests); err != nil {
		return 0, err
	}
	return synced, nil
}

func (s *APIService) GetRegistryVersion(ctx context.Context) (*outport.RegistryVersion, error) {
	return s.registry.GetLatest(ctx)
}
