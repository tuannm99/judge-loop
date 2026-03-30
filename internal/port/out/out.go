package out

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/tuannm99/judge-loop/internal/domain"
	"github.com/tuannm99/judge-loop/internal/infrastructure/sandbox"
)

type ProblemFilter struct {
	Difficulty *domain.Difficulty
	Tag        string
	Pattern    string
	Provider   *domain.Provider
	Limit      int
	Offset     int
}

type StreakInfo struct {
	Current       int
	Longest       int
	LastPracticed *time.Time
}

type DueReview struct {
	ProblemID   uuid.UUID
	Slug        string
	Title       string
	LastSolved  *time.Time
	DaysOverdue int
}

type RegistryVersion struct {
	Version   string
	UpdatedAt time.Time
	SyncedAt  time.Time
}

type EvaluateSubmissionJob struct {
	SubmissionID string
	UserID       string
}

type RunRequest struct {
	Language string
	Code     string
	Input    string
}

type ProblemRepository interface {
	List(ctx context.Context, filter ProblemFilter) ([]domain.Problem, int, error)
	GetByID(ctx context.Context, id uuid.UUID) (*domain.Problem, error)
	GetBySlug(ctx context.Context, slug string) (*domain.Problem, error)
	Suggest(ctx context.Context, userID uuid.UUID, patterns []string) (*domain.Problem, error)
	UpsertFromManifest(ctx context.Context, m domain.ProblemManifest) error
}

type SubmissionRepository interface {
	Create(ctx context.Context, sub *domain.Submission) error
	GetByID(ctx context.Context, id uuid.UUID) (*domain.Submission, error)
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

type SessionRepository interface {
	GetOrCreateToday(ctx context.Context, userID uuid.UUID) (*domain.DailySession, error)
	RecordSubmission(ctx context.Context, userID uuid.UUID, accepted bool) error
	StartTimer(ctx context.Context, userID uuid.UUID, problemID *uuid.UUID) (*domain.TimerSession, error)
	StopTimer(ctx context.Context, userID uuid.UUID) (*domain.TimerSession, error)
	ActiveTimer(ctx context.Context, userID uuid.UUID) (*domain.TimerSession, error)
	GetStreak(ctx context.Context, userID uuid.UUID) (StreakInfo, error)
}

type ReviewRepository interface {
	GetDue(ctx context.Context, userID uuid.UUID) ([]DueReview, error)
	Upsert(ctx context.Context, userID, problemID uuid.UUID) error
}

type RegistryRepository interface {
	GetLatest(ctx context.Context) (*RegistryVersion, error)
	Save(ctx context.Context, version string, updatedAt time.Time, refs []domain.ManifestRef) error
}

type TestCaseRepository interface {
	GetByProblem(ctx context.Context, problemID uuid.UUID) ([]domain.TestCase, error)
}

type EvaluationPublisher interface {
	PublishEvaluation(job EvaluateSubmissionJob) error
}

type CodeRunner interface {
	Run(ctx context.Context, req RunRequest) (sandbox.RunResult, error)
}
