package mocks

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"github.com/tuannm99/judge-loop/internal/domain"
	inport "github.com/tuannm99/judge-loop/internal/port/in"
	outport "github.com/tuannm99/judge-loop/internal/port/out"
)

func TestMockAPIService(t *testing.T) {
	ctx := context.Background()
	userID := uuid.New()
	problemID := uuid.New()
	submissionID := uuid.New()
	sessionID := uuid.New()
	now := time.Now()
	sub := &domain.Submission{ID: submissionID}
	problem := &domain.Problem{ID: problemID}
	timer := &domain.TimerSession{ID: sessionID}
	progress := inport.ProgressToday{Solved: 1}
	streak := outport.StreakInfo{Current: 2}
	reviews := []outport.DueReview{{ProblemID: problemID}}
	registry := &outport.RegistryVersion{Version: "v1"}
	manifests := []domain.ManifestRef{{Name: "main", Path: "providers/main.json"}}
	problems := []domain.Problem{{ID: problemID}}
	problemList := []domain.ProblemManifest{{Slug: "two-sum"}}
	subs := []domain.Submission{{ID: submissionID}}

	m := NewMockAPIService(t)
	require.NotNil(t, m.EXPECT())

	m.EXPECT().CreateSubmission(mock.Anything, userID, problemID, "go", "code", &sessionID).
		Run(func(ctx context.Context, gotUserID uuid.UUID, gotProblemID uuid.UUID, language, code string, gotSessionID *uuid.UUID) {
			require.Equal(t, userID, gotUserID)
			require.Equal(t, sessionID, *gotSessionID)
		}).
		Return(sub, nil)
	gotSub, err := m.CreateSubmission(ctx, userID, problemID, "go", "code", &sessionID)
	require.NoError(t, err)
	require.Equal(t, sub, gotSub)
	m.EXPECT().CreateSubmission(mock.Anything, userID, problemID, "py", "code", (*uuid.UUID)(nil)).
		RunAndReturn(func(context.Context, uuid.UUID, uuid.UUID, string, string, *uuid.UUID) (*domain.Submission, error) {
			return sub, nil
		})
	_, _ = m.CreateSubmission(ctx, userID, problemID, "py", "code", nil)

	m.EXPECT().CurrentTimer(mock.Anything, userID).Run(func(context.Context, uuid.UUID) {}).Return(timer, nil)
	_, _ = m.CurrentTimer(ctx, userID)
	m.EXPECT().CurrentTimer(mock.Anything, userID).RunAndReturn(func(context.Context, uuid.UUID) (*domain.TimerSession, error) { return timer, nil })
	_, _ = m.CurrentTimer(ctx, userID)

	m.EXPECT().GetProblem(mock.Anything, "slug").Run(func(context.Context, string) {}).Return(problem, nil)
	_, _ = m.GetProblem(ctx, "slug")
	m.EXPECT().GetProblem(mock.Anything, "slug2").RunAndReturn(func(context.Context, string) (*domain.Problem, error) { return problem, nil })
	_, _ = m.GetProblem(ctx, "slug2")

	m.EXPECT().GetProgressToday(mock.Anything, userID).Run(func(context.Context, uuid.UUID) {}).Return(progress, nil)
	_, _ = m.GetProgressToday(ctx, userID)
	m.EXPECT().GetProgressToday(mock.Anything, userID).RunAndReturn(func(context.Context, uuid.UUID) (inport.ProgressToday, error) { return progress, nil })
	_, _ = m.GetProgressToday(ctx, userID)

	m.EXPECT().GetRegistryVersion(mock.Anything).Run(func(context.Context) {}).Return(registry, nil)
	_, _ = m.GetRegistryVersion(ctx)
	m.EXPECT().GetRegistryVersion(mock.Anything).RunAndReturn(func(context.Context) (*outport.RegistryVersion, error) { return registry, nil })
	_, _ = m.GetRegistryVersion(ctx)

	m.EXPECT().GetReviewsToday(mock.Anything, userID).Run(func(context.Context, uuid.UUID) {}).Return(reviews, nil)
	_, _ = m.GetReviewsToday(ctx, userID)
	m.EXPECT().GetReviewsToday(mock.Anything, userID).RunAndReturn(func(context.Context, uuid.UUID) ([]outport.DueReview, error) { return reviews, nil })
	_, _ = m.GetReviewsToday(ctx, userID)

	m.EXPECT().GetStreak(mock.Anything, userID).Run(func(context.Context, uuid.UUID) {}).Return(streak, nil)
	_, _ = m.GetStreak(ctx, userID)
	m.EXPECT().GetStreak(mock.Anything, userID).RunAndReturn(func(context.Context, uuid.UUID) (outport.StreakInfo, error) { return streak, nil })
	_, _ = m.GetStreak(ctx, userID)

	m.EXPECT().GetSubmission(mock.Anything, submissionID).Run(func(context.Context, uuid.UUID) {}).Return(sub, nil)
	_, _ = m.GetSubmission(ctx, submissionID)
	m.EXPECT().GetSubmission(mock.Anything, submissionID).RunAndReturn(func(context.Context, uuid.UUID) (*domain.Submission, error) { return sub, nil })
	_, _ = m.GetSubmission(ctx, submissionID)

	filter := outport.ProblemFilter{Limit: 1}
	m.EXPECT().ListProblems(mock.Anything, filter).Run(func(context.Context, outport.ProblemFilter) {}).Return(problems, 1, nil)
	_, _, _ = m.ListProblems(ctx, filter)
	m.EXPECT().ListProblems(mock.Anything, filter).RunAndReturn(func(context.Context, outport.ProblemFilter) ([]domain.Problem, int, error) { return problems, 1, nil })
	_, _, _ = m.ListProblems(ctx, filter)

	m.EXPECT().ListSubmissions(mock.Anything, userID, (*uuid.UUID)(nil), 1, 0).
		Run(func(context.Context, uuid.UUID, *uuid.UUID, int, int) {}).
		Return(subs, nil)
	_, _ = m.ListSubmissions(ctx, userID, nil, 1, 0)
	m.EXPECT().ListSubmissions(mock.Anything, userID, (*uuid.UUID)(nil), 1, 0).
		RunAndReturn(func(context.Context, uuid.UUID, *uuid.UUID, int, int) ([]domain.Submission, error) { return subs, nil })
	_, _ = m.ListSubmissions(ctx, userID, nil, 1, 0)

	m.EXPECT().StartTimer(mock.Anything, userID, (*uuid.UUID)(nil)).Run(func(context.Context, uuid.UUID, *uuid.UUID) {}).Return(timer, nil)
	_, _ = m.StartTimer(ctx, userID, nil)
	m.EXPECT().StartTimer(mock.Anything, userID, (*uuid.UUID)(nil)).RunAndReturn(func(context.Context, uuid.UUID, *uuid.UUID) (*domain.TimerSession, error) { return timer, nil })
	_, _ = m.StartTimer(ctx, userID, nil)

	m.EXPECT().StopTimer(mock.Anything, userID).Run(func(context.Context, uuid.UUID) {}).Return(timer, nil)
	_, _ = m.StopTimer(ctx, userID)
	m.EXPECT().StopTimer(mock.Anything, userID).RunAndReturn(func(context.Context, uuid.UUID) (*domain.TimerSession, error) { return timer, nil })
	_, _ = m.StopTimer(ctx, userID)

	m.EXPECT().SuggestProblem(mock.Anything, userID).Run(func(context.Context, uuid.UUID) {}).Return(problem, nil)
	_, _ = m.SuggestProblem(ctx, userID)
	m.EXPECT().SuggestProblem(mock.Anything, userID).RunAndReturn(func(context.Context, uuid.UUID) (*domain.Problem, error) { return problem, nil })
	_, _ = m.SuggestProblem(ctx, userID)

	m.EXPECT().SyncRegistry(mock.Anything, "v1", now, problemList, manifests).
		Run(func(context.Context, string, time.Time, []domain.ProblemManifest, []domain.ManifestRef) {}).
		Return(1, nil)
	_, _ = m.SyncRegistry(ctx, "v1", now, problemList, manifests)
	m.EXPECT().SyncRegistry(mock.Anything, "v2", now, problemList, manifests).
		RunAndReturn(func(context.Context, string, time.Time, []domain.ProblemManifest, []domain.ManifestRef) (int, error) {
			return 1, nil
		})
	_, _ = m.SyncRegistry(ctx, "v2", now, problemList, manifests)
}

func TestMockEvaluationService(t *testing.T) {
	ctx := context.Background()
	subID := uuid.New()
	userID := uuid.New()

	m := NewMockEvaluationService(t)
	require.NotNil(t, m.EXPECT())
	m.EXPECT().EvaluateSubmission(mock.Anything, subID, userID, 2).Run(func(context.Context, uuid.UUID, uuid.UUID, int) {}).Return(nil)
	require.NoError(t, m.EvaluateSubmission(ctx, subID, userID, 2))
	m.EXPECT().EvaluateSubmission(mock.Anything, subID, userID, 3).RunAndReturn(func(context.Context, uuid.UUID, uuid.UUID, int) error { return nil })
	require.NoError(t, m.EvaluateSubmission(ctx, subID, userID, 3))
}
