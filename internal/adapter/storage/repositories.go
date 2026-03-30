package storageadapter

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/tuannm99/judge-loop/internal/domain"
	postgres "github.com/tuannm99/judge-loop/internal/infrastructure/postgres"
	outport "github.com/tuannm99/judge-loop/internal/port/out"
)

type ProblemRepository struct{ store *postgres.ProblemStore }

func NewProblemRepository(db *postgres.DB) *ProblemRepository {
	return &ProblemRepository{store: postgres.NewProblemStore(db)}
}

func (r *ProblemRepository) List(ctx context.Context, filter outport.ProblemFilter) ([]domain.Problem, int, error) {
	return r.store.List(ctx, postgres.ProblemFilter{
		Difficulty: filter.Difficulty,
		Tag:        filter.Tag,
		Pattern:    filter.Pattern,
		Provider:   filter.Provider,
		Limit:      filter.Limit,
		Offset:     filter.Offset,
	})
}

func (r *ProblemRepository) GetByID(ctx context.Context, id uuid.UUID) (*domain.Problem, error) {
	return r.store.GetByID(ctx, id)
}

func (r *ProblemRepository) GetBySlug(ctx context.Context, slug string) (*domain.Problem, error) {
	return r.store.GetBySlug(ctx, slug)
}

func (r *ProblemRepository) Suggest(ctx context.Context, userID uuid.UUID, patterns []string) (*domain.Problem, error) {
	return r.store.Suggest(ctx, userID, patterns)
}

func (r *ProblemRepository) UpsertFromManifest(ctx context.Context, m domain.ProblemManifest) error {
	return r.store.UpsertFromManifest(ctx, m)
}

type SubmissionRepository struct{ store *postgres.SubmissionStore }

func NewSubmissionRepository(db *postgres.DB) *SubmissionRepository {
	return &SubmissionRepository{store: postgres.NewSubmissionStore(db)}
}

func (r *SubmissionRepository) Create(ctx context.Context, sub *domain.Submission) error {
	return r.store.Create(ctx, sub)
}

func (r *SubmissionRepository) GetByID(ctx context.Context, id uuid.UUID) (*domain.Submission, error) {
	return r.store.GetByID(ctx, id)
}

func (r *SubmissionRepository) UpdateVerdict(
	ctx context.Context,
	id uuid.UUID,
	status, verdict string,
	passed, total int,
	runtimeMS int64,
	errMsg string,
	evaluatedAt *time.Time,
) error {
	return r.store.UpdateVerdict(ctx, id, status, verdict, passed, total, runtimeMS, errMsg, evaluatedAt)
}

func (r *SubmissionRepository) ListByUser(
	ctx context.Context,
	userID uuid.UUID,
	problemID *uuid.UUID,
	limit, offset int,
) ([]domain.Submission, error) {
	return r.store.ListByUser(ctx, userID, problemID, limit, offset)
}

type SessionRepository struct{ store *postgres.SessionStore }

func NewSessionRepository(db *postgres.DB) *SessionRepository {
	return &SessionRepository{store: postgres.NewSessionStore(db)}
}

func (r *SessionRepository) GetOrCreateToday(ctx context.Context, userID uuid.UUID) (*domain.DailySession, error) {
	return r.store.GetOrCreateToday(ctx, userID)
}

func (r *SessionRepository) RecordSubmission(ctx context.Context, userID uuid.UUID, accepted bool) error {
	return r.store.RecordSubmission(ctx, userID, accepted)
}

func (r *SessionRepository) StartTimer(
	ctx context.Context,
	userID uuid.UUID,
	problemID *uuid.UUID,
) (*domain.TimerSession, error) {
	return r.store.StartTimer(ctx, userID, problemID)
}

func (r *SessionRepository) StopTimer(ctx context.Context, userID uuid.UUID) (*domain.TimerSession, error) {
	return r.store.StopTimer(ctx, userID)
}

func (r *SessionRepository) ActiveTimer(ctx context.Context, userID uuid.UUID) (*domain.TimerSession, error) {
	return r.store.ActiveTimer(ctx, userID)
}

func (r *SessionRepository) GetStreak(ctx context.Context, userID uuid.UUID) (outport.StreakInfo, error) {
	info, err := r.store.GetStreak(ctx, userID)
	if err != nil {
		return outport.StreakInfo{}, err
	}
	return outport.StreakInfo(info), nil
}

type ReviewRepository struct{ store *postgres.ReviewStore }

func NewReviewRepository(db *postgres.DB) *ReviewRepository {
	return &ReviewRepository{store: postgres.NewReviewStore(db)}
}

func (r *ReviewRepository) GetDue(ctx context.Context, userID uuid.UUID) ([]outport.DueReview, error) {
	reviews, err := r.store.GetDue(ctx, userID)
	if err != nil {
		return nil, err
	}
	out := make([]outport.DueReview, 0, len(reviews))
	for _, review := range reviews {
		out = append(out, outport.DueReview(review))
	}
	return out, nil
}

func (r *ReviewRepository) Upsert(ctx context.Context, userID, problemID uuid.UUID) error {
	return r.store.Upsert(ctx, userID, problemID)
}

type RegistryRepository struct{ store *postgres.RegistryStore }

func NewRegistryRepository(db *postgres.DB) *RegistryRepository {
	return &RegistryRepository{store: postgres.NewRegistryStore(db)}
}

func (r *RegistryRepository) GetLatest(ctx context.Context) (*outport.RegistryVersion, error) {
	row, err := r.store.GetLatest(ctx)
	if err != nil || row == nil {
		if err != nil {
			return nil, err
		}
		return nil, nil
	}
	return &outport.RegistryVersion{
		Version:   row.Version,
		UpdatedAt: row.UpdatedAt,
		SyncedAt:  row.SyncedAt,
	}, nil
}

func (r *RegistryRepository) Save(
	ctx context.Context,
	version string,
	updatedAt time.Time,
	refs []domain.ManifestRef,
) error {
	return r.store.Save(ctx, version, updatedAt, refs)
}

type TestCaseRepository struct{ store *postgres.TestCaseStore }

func NewTestCaseRepository(db *postgres.DB) *TestCaseRepository {
	return &TestCaseRepository{store: postgres.NewTestCaseStore(db)}
}

func (r *TestCaseRepository) GetByProblem(ctx context.Context, problemID uuid.UUID) ([]domain.TestCase, error) {
	return r.store.GetByProblem(ctx, problemID)
}
