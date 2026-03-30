package in

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/tuannm99/judge-loop/internal/domain"
	out "github.com/tuannm99/judge-loop/internal/port/out"
)

type APIService interface {
	ListProblems(ctx context.Context, filter out.ProblemFilter) ([]domain.Problem, int, error)
	GetProblem(ctx context.Context, rawID string) (*domain.Problem, error)
	SuggestProblem(ctx context.Context, userID uuid.UUID) (*domain.Problem, error)
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
	GetProgressToday(ctx context.Context, userID uuid.UUID) (ProgressToday, error)
	GetStreak(ctx context.Context, userID uuid.UUID) (out.StreakInfo, error)
	StartTimer(ctx context.Context, userID uuid.UUID, problemID *uuid.UUID) (*domain.TimerSession, error)
	StopTimer(ctx context.Context, userID uuid.UUID) (*domain.TimerSession, error)
	CurrentTimer(ctx context.Context, userID uuid.UUID) (*domain.TimerSession, error)
	GetReviewsToday(ctx context.Context, userID uuid.UUID) ([]out.DueReview, error)
	SyncRegistry(
		ctx context.Context,
		version string,
		updatedAt time.Time,
		problems []domain.ProblemManifest,
		manifests []domain.ManifestRef,
	) (int, error)
	GetRegistryVersion(ctx context.Context) (*out.RegistryVersion, error)
}

type EvaluationService interface {
	EvaluateSubmission(ctx context.Context, submissionID, userID uuid.UUID, timeLimitSecs int) error
}

type ProgressToday struct {
	Date             string
	Solved           int
	Attempted        int
	TimeSpentMinutes int
	Streak           int
}
