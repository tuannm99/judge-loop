package mocks

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"github.com/tuannm99/judge-loop/internal/domain"
	"github.com/tuannm99/judge-loop/internal/domain/judge"
	outport "github.com/tuannm99/judge-loop/internal/port/out"
)

func TestRepositoryAndRunnerMocks(t *testing.T) {
	ctx := context.Background()
	userID := uuid.New()
	problemID := uuid.New()
	subID := uuid.New()
	sessionID := uuid.New()
	now := time.Now()
	problem := &domain.Problem{ID: problemID}
	sub := &domain.Submission{ID: subID}
	timer := &domain.TimerSession{ID: sessionID}
	testCases := []domain.TestCase{{ProblemID: problemID}}
	problems := []domain.Problem{{ID: problemID}}
	subs := []domain.Submission{{ID: subID}}
	reviews := []outport.DueReview{{ProblemID: problemID}}
	streak := outport.StreakInfo{Current: 2}
	registry := &outport.RegistryVersion{Version: "v1"}
	runReq := outport.RunRequest{Language: "go", Code: "code", Input: "1"}
	runRes := judge.RunResult{Output: "1"}

	problemRepo := NewMockProblemRepository(t)
	require.NotNil(t, problemRepo.EXPECT())
	filter := outport.ProblemFilter{Limit: 1}
	problemRepo.EXPECT().List(mock.Anything, filter).Run(func(context.Context, outport.ProblemFilter) {}).Return(problems, 1, nil)
	_, _, _ = problemRepo.List(ctx, filter)
	problemRepo.EXPECT().List(mock.Anything, filter).RunAndReturn(func(context.Context, outport.ProblemFilter) ([]domain.Problem, int, error) { return problems, 1, nil })
	_, _, _ = problemRepo.List(ctx, filter)
	problemRepo.EXPECT().GetByID(mock.Anything, problemID).Run(func(context.Context, uuid.UUID) {}).Return(problem, nil)
	_, _ = problemRepo.GetByID(ctx, problemID)
	problemRepo.EXPECT().GetByID(mock.Anything, problemID).RunAndReturn(func(context.Context, uuid.UUID) (*domain.Problem, error) { return problem, nil })
	_, _ = problemRepo.GetByID(ctx, problemID)
	problemRepo.EXPECT().GetBySlug(mock.Anything, "slug").Run(func(context.Context, string) {}).Return(problem, nil)
	_, _ = problemRepo.GetBySlug(ctx, "slug")
	problemRepo.EXPECT().GetBySlug(mock.Anything, "slug2").RunAndReturn(func(context.Context, string) (*domain.Problem, error) { return problem, nil })
	_, _ = problemRepo.GetBySlug(ctx, "slug2")
	problemRepo.EXPECT().Suggest(mock.Anything, userID, []string{"dp"}).Run(func(context.Context, uuid.UUID, []string) {}).Return(problem, nil)
	_, _ = problemRepo.Suggest(ctx, userID, []string{"dp"})
	problemRepo.EXPECT().Suggest(mock.Anything, userID, []string{"dp"}).RunAndReturn(func(context.Context, uuid.UUID, []string) (*domain.Problem, error) { return problem, nil })
	_, _ = problemRepo.Suggest(ctx, userID, []string{"dp"})
	manifest := domain.ProblemManifest{Slug: "two-sum"}
	problemRepo.EXPECT().UpsertFromManifest(mock.Anything, manifest).Run(func(context.Context, domain.ProblemManifest) {}).Return(nil)
	require.NoError(t, problemRepo.UpsertFromManifest(ctx, manifest))
	problemRepo.EXPECT().UpsertFromManifest(mock.Anything, manifest).RunAndReturn(func(context.Context, domain.ProblemManifest) error { return nil })
	require.NoError(t, problemRepo.UpsertFromManifest(ctx, manifest))

	submissionRepo := NewMockSubmissionRepository(t)
	require.NotNil(t, submissionRepo.EXPECT())
	submissionRepo.EXPECT().Create(mock.Anything, sub).Run(func(context.Context, *domain.Submission) {}).Return(nil)
	require.NoError(t, submissionRepo.Create(ctx, sub))
	submissionRepo.EXPECT().Create(mock.Anything, sub).RunAndReturn(func(context.Context, *domain.Submission) error { return nil })
	require.NoError(t, submissionRepo.Create(ctx, sub))
	submissionRepo.EXPECT().GetByID(mock.Anything, subID).Run(func(context.Context, uuid.UUID) {}).Return(sub, nil)
	_, _ = submissionRepo.GetByID(ctx, subID)
	submissionRepo.EXPECT().GetByID(mock.Anything, subID).RunAndReturn(func(context.Context, uuid.UUID) (*domain.Submission, error) { return sub, nil })
	_, _ = submissionRepo.GetByID(ctx, subID)
	submissionRepo.EXPECT().ListByUser(mock.Anything, userID, (*uuid.UUID)(nil), 1, 0).Run(func(context.Context, uuid.UUID, *uuid.UUID, int, int) {}).Return(subs, nil)
	_, _ = submissionRepo.ListByUser(ctx, userID, nil, 1, 0)
	submissionRepo.EXPECT().ListByUser(mock.Anything, userID, (*uuid.UUID)(nil), 1, 0).RunAndReturn(func(context.Context, uuid.UUID, *uuid.UUID, int, int) ([]domain.Submission, error) { return subs, nil })
	_, _ = submissionRepo.ListByUser(ctx, userID, nil, 1, 0)
	submissionRepo.EXPECT().UpdateVerdict(mock.Anything, subID, "accepted", "Accepted", 1, 1, int64(2), "", &now).Run(func(context.Context, uuid.UUID, string, string, int, int, int64, string, *time.Time) {}).Return(nil)
	require.NoError(t, submissionRepo.UpdateVerdict(ctx, subID, "accepted", "Accepted", 1, 1, 2, "", &now))
	submissionRepo.EXPECT().UpdateVerdict(mock.Anything, subID, "accepted", "Accepted", 1, 1, int64(2), "", &now).RunAndReturn(func(context.Context, uuid.UUID, string, string, int, int, int64, string, *time.Time) error {
		return nil
	})
	require.NoError(t, submissionRepo.UpdateVerdict(ctx, subID, "accepted", "Accepted", 1, 1, 2, "", &now))

	sessionRepo := NewMockSessionRepository(t)
	require.NotNil(t, sessionRepo.EXPECT())
	sessionRepo.EXPECT().ActiveTimer(mock.Anything, userID).Run(func(context.Context, uuid.UUID) {}).Return(timer, nil)
	_, _ = sessionRepo.ActiveTimer(ctx, userID)
	sessionRepo.EXPECT().ActiveTimer(mock.Anything, userID).RunAndReturn(func(context.Context, uuid.UUID) (*domain.TimerSession, error) { return timer, nil })
	_, _ = sessionRepo.ActiveTimer(ctx, userID)
	daily := &domain.DailySession{UserID: userID}
	sessionRepo.EXPECT().GetOrCreateToday(mock.Anything, userID).Run(func(context.Context, uuid.UUID) {}).Return(daily, nil)
	_, _ = sessionRepo.GetOrCreateToday(ctx, userID)
	sessionRepo.EXPECT().GetOrCreateToday(mock.Anything, userID).RunAndReturn(func(context.Context, uuid.UUID) (*domain.DailySession, error) { return daily, nil })
	_, _ = sessionRepo.GetOrCreateToday(ctx, userID)
	sessionRepo.EXPECT().GetStreak(mock.Anything, userID).Run(func(context.Context, uuid.UUID) {}).Return(streak, nil)
	_, _ = sessionRepo.GetStreak(ctx, userID)
	sessionRepo.EXPECT().GetStreak(mock.Anything, userID).RunAndReturn(func(context.Context, uuid.UUID) (outport.StreakInfo, error) { return streak, nil })
	_, _ = sessionRepo.GetStreak(ctx, userID)
	sessionRepo.EXPECT().RecordSubmission(mock.Anything, userID, true).Run(func(context.Context, uuid.UUID, bool) {}).Return(nil)
	require.NoError(t, sessionRepo.RecordSubmission(ctx, userID, true))
	sessionRepo.EXPECT().RecordSubmission(mock.Anything, userID, true).RunAndReturn(func(context.Context, uuid.UUID, bool) error { return nil })
	require.NoError(t, sessionRepo.RecordSubmission(ctx, userID, true))
	sessionRepo.EXPECT().StartTimer(mock.Anything, userID, (*uuid.UUID)(nil)).Run(func(context.Context, uuid.UUID, *uuid.UUID) {}).Return(timer, nil)
	_, _ = sessionRepo.StartTimer(ctx, userID, nil)
	sessionRepo.EXPECT().StartTimer(mock.Anything, userID, (*uuid.UUID)(nil)).RunAndReturn(func(context.Context, uuid.UUID, *uuid.UUID) (*domain.TimerSession, error) { return timer, nil })
	_, _ = sessionRepo.StartTimer(ctx, userID, nil)
	sessionRepo.EXPECT().StopTimer(mock.Anything, userID).Run(func(context.Context, uuid.UUID) {}).Return(timer, nil)
	_, _ = sessionRepo.StopTimer(ctx, userID)
	sessionRepo.EXPECT().StopTimer(mock.Anything, userID).RunAndReturn(func(context.Context, uuid.UUID) (*domain.TimerSession, error) { return timer, nil })
	_, _ = sessionRepo.StopTimer(ctx, userID)

	reviewRepo := NewMockReviewRepository(t)
	require.NotNil(t, reviewRepo.EXPECT())
	reviewRepo.EXPECT().GetDue(mock.Anything, userID).Run(func(context.Context, uuid.UUID) {}).Return(reviews, nil)
	_, _ = reviewRepo.GetDue(ctx, userID)
	reviewRepo.EXPECT().GetDue(mock.Anything, userID).RunAndReturn(func(context.Context, uuid.UUID) ([]outport.DueReview, error) { return reviews, nil })
	_, _ = reviewRepo.GetDue(ctx, userID)
	reviewRepo.EXPECT().Upsert(mock.Anything, userID, problemID).Run(func(context.Context, uuid.UUID, uuid.UUID) {}).Return(nil)
	require.NoError(t, reviewRepo.Upsert(ctx, userID, problemID))
	reviewRepo.EXPECT().Upsert(mock.Anything, userID, problemID).RunAndReturn(func(context.Context, uuid.UUID, uuid.UUID) error { return nil })
	require.NoError(t, reviewRepo.Upsert(ctx, userID, problemID))

	registryRepo := NewMockRegistryRepository(t)
	require.NotNil(t, registryRepo.EXPECT())
	registryRepo.EXPECT().GetLatest(mock.Anything).Run(func(context.Context) {}).Return(registry, nil)
	_, _ = registryRepo.GetLatest(ctx)
	registryRepo.EXPECT().GetLatest(mock.Anything).RunAndReturn(func(context.Context) (*outport.RegistryVersion, error) { return registry, nil })
	_, _ = registryRepo.GetLatest(ctx)
	refs := []domain.ManifestRef{{Name: "main", Path: "providers/main.json"}}
	registryRepo.EXPECT().Save(mock.Anything, "v1", now, refs).Run(func(context.Context, string, time.Time, []domain.ManifestRef) {}).Return(nil)
	require.NoError(t, registryRepo.Save(ctx, "v1", now, refs))
	registryRepo.EXPECT().Save(mock.Anything, "v1", now, refs).RunAndReturn(func(context.Context, string, time.Time, []domain.ManifestRef) error { return nil })
	require.NoError(t, registryRepo.Save(ctx, "v1", now, refs))

	testCaseRepo := NewMockTestCaseRepository(t)
	require.NotNil(t, testCaseRepo.EXPECT())
	testCaseRepo.EXPECT().GetByProblem(mock.Anything, problemID).Run(func(context.Context, uuid.UUID) {}).Return(testCases, nil)
	_, _ = testCaseRepo.GetByProblem(ctx, problemID)
	testCaseRepo.EXPECT().GetByProblem(mock.Anything, problemID).RunAndReturn(func(context.Context, uuid.UUID) ([]domain.TestCase, error) { return testCases, nil })
	_, _ = testCaseRepo.GetByProblem(ctx, problemID)

	publisher := NewMockEvaluationPublisher(t)
	require.NotNil(t, publisher.EXPECT())
	job := outport.EvaluateSubmissionJob{SubmissionID: subID.String(), UserID: userID.String()}
	publisher.EXPECT().PublishEvaluation(job).Run(func(outport.EvaluateSubmissionJob) {}).Return(nil)
	require.NoError(t, publisher.PublishEvaluation(job))
	publisher.EXPECT().PublishEvaluation(job).RunAndReturn(func(outport.EvaluateSubmissionJob) error { return nil })
	require.NoError(t, publisher.PublishEvaluation(job))

	runner := NewMockCodeRunner(t)
	require.NotNil(t, runner.EXPECT())
	runner.EXPECT().Run(mock.Anything, runReq).Run(func(context.Context, outport.RunRequest) {}).Return(runRes, nil)
	_, _ = runner.Run(ctx, runReq)
	runner.EXPECT().Run(mock.Anything, runReq).RunAndReturn(func(context.Context, outport.RunRequest) (judge.RunResult, error) { return runRes, nil })
	_, _ = runner.Run(ctx, runReq)
}
