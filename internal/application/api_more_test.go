package application

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"github.com/tuannm99/judge-loop/internal/domain"
	inport "github.com/tuannm99/judge-loop/internal/port/in"
	outport "github.com/tuannm99/judge-loop/internal/port/out"
	outmocks "github.com/tuannm99/judge-loop/internal/port/out/mocks"
)

func TestAPIServiceDelegatesRemainingMethods(t *testing.T) {
	problems := outmocks.NewMockProblemRepository(t)
	submissions := outmocks.NewMockSubmissionRepository(t)
	sessions := outmocks.NewMockSessionRepository(t)
	reviews := outmocks.NewMockReviewRepository(t)
	registry := outmocks.NewMockRegistryRepository(t)
	publisher := outmocks.NewMockEvaluationPublisher(t)

	svc := NewAPIService(problems, submissions, sessions, reviews, registry, publisher)
	ctx := context.Background()
	userID := uuid.New()
	problemID := uuid.New()
	submissionID := uuid.New()

	diff := domain.DifficultyEasy
	filter := outport.ProblemFilter{Difficulty: &diff, Limit: 10}
	expectedProblems := []domain.Problem{{ID: problemID}}
	problems.EXPECT().List(ctx, filter).Return(expectedProblems, 1, nil)
	gotProblems, total, err := svc.ListProblems(ctx, filter)
	require.NoError(t, err)
	require.Equal(t, expectedProblems, gotProblems)
	require.Equal(t, 1, total)

	problem := &domain.Problem{ID: problemID}
	problems.EXPECT().GetByID(ctx, problemID).Return(problem, nil)
	gotProblem, err := svc.GetProblem(ctx, problemID.String())
	require.NoError(t, err)
	require.Equal(t, problem, gotProblem)

	problems.EXPECT().Suggest(ctx, userID, []string(nil)).Return(problem, nil)
	gotProblem, err = svc.SuggestProblem(ctx, userID)
	require.NoError(t, err)
	require.Equal(t, problem, gotProblem)

	sub := &domain.Submission{ID: submissionID}
	submissions.EXPECT().GetByID(ctx, submissionID).Return(sub, nil)
	gotSub, err := svc.GetSubmission(ctx, submissionID)
	require.NoError(t, err)
	require.Equal(t, sub, gotSub)

	subList := []domain.Submission{{ID: submissionID}}
	problemPtr := &problemID
	submissions.EXPECT().ListByUser(ctx, userID, problemPtr, 20, 5).Return(subList, nil)
	gotSubs, err := svc.ListSubmissions(ctx, userID, problemPtr, 20, 5)
	require.NoError(t, err)
	require.Equal(t, subList, gotSubs)

	ds := &domain.DailySession{
		Date:           time.Date(2026, 3, 31, 0, 0, 0, 0, time.UTC),
		SolvedCount:    2,
		AttemptedCount: 3,
		TimeSpentSecs:  125,
	}
	streak := outport.StreakInfo{Current: 4}
	sessions.EXPECT().GetOrCreateToday(ctx, userID).Return(ds, nil)
	sessions.EXPECT().GetStreak(ctx, userID).Return(streak, nil)
	progress, err := svc.GetProgressToday(ctx, userID)
	require.NoError(t, err)
	require.Equal(t, inport.ProgressToday{
		Date:             "2026-03-31",
		Solved:           2,
		Attempted:        3,
		TimeSpentMinutes: 2,
		Streak:           4,
	}, progress)

	sessions.EXPECT().GetStreak(ctx, userID).Return(streak, nil)
	gotStreak, err := svc.GetStreak(ctx, userID)
	require.NoError(t, err)
	require.Equal(t, streak, gotStreak)

	timer := &domain.TimerSession{ID: uuid.New()}
	sessions.EXPECT().StartTimer(ctx, userID, problemPtr).Return(timer, nil)
	gotTimer, err := svc.StartTimer(ctx, userID, problemPtr)
	require.NoError(t, err)
	require.Equal(t, timer, gotTimer)

	sessions.EXPECT().StopTimer(ctx, userID).Return(timer, nil)
	gotTimer, err = svc.StopTimer(ctx, userID)
	require.NoError(t, err)
	require.Equal(t, timer, gotTimer)

	sessions.EXPECT().ActiveTimer(ctx, userID).Return(timer, nil)
	gotTimer, err = svc.CurrentTimer(ctx, userID)
	require.NoError(t, err)
	require.Equal(t, timer, gotTimer)

	due := []outport.DueReview{{ProblemID: problemID}}
	reviews.EXPECT().GetDue(ctx, userID).Return(due, nil)
	gotDue, err := svc.GetReviewsToday(ctx, userID)
	require.NoError(t, err)
	require.Equal(t, due, gotDue)

	version := &outport.RegistryVersion{Version: "v1"}
	registry.EXPECT().GetLatest(ctx).Return(version, nil)
	gotVersion, err := svc.GetRegistryVersion(ctx)
	require.NoError(t, err)
	require.Equal(t, version, gotVersion)
}

func TestAPIServiceProgressAndRegistryErrors(t *testing.T) {
	problems := outmocks.NewMockProblemRepository(t)
	submissions := outmocks.NewMockSubmissionRepository(t)
	sessions := outmocks.NewMockSessionRepository(t)
	reviews := outmocks.NewMockReviewRepository(t)
	registry := outmocks.NewMockRegistryRepository(t)

	svc := NewAPIService(problems, submissions, sessions, reviews, registry, nil)
	ctx := context.Background()
	userID := uuid.New()

	sessions.EXPECT().GetOrCreateToday(ctx, userID).Return(nil, errors.New("daily"))
	_, err := svc.GetProgressToday(ctx, userID)
	require.Error(t, err)

	submissions.EXPECT().Create(ctx, mock.AnythingOfType("*domain.Submission")).Return(errors.New("create"))
	_, err = svc.CreateSubmission(ctx, userID, uuid.New(), "go", "code", nil)
	require.Error(t, err)

	sessions2 := outmocks.NewMockSessionRepository(t)
	svc2 := NewAPIService(problems, submissions, sessions2, reviews, registry, nil)
	sessions2.EXPECT().GetOrCreateToday(ctx, userID).Return(&domain.DailySession{}, nil)
	sessions2.EXPECT().GetStreak(ctx, userID).Return(outport.StreakInfo{}, errors.New("streak"))
	_, err = svc2.GetProgressToday(ctx, userID)
	require.Error(t, err)

	manifest := domain.ProblemManifest{Slug: "two-sum"}
	problems.EXPECT().UpsertFromManifest(ctx, manifest).Return(errors.New("upsert"))
	_, err = svc.SyncRegistry(ctx, "v1", time.Time{}, []domain.ProblemManifest{manifest}, nil)
	require.Error(t, err)

	problems2 := outmocks.NewMockProblemRepository(t)
	registry2 := outmocks.NewMockRegistryRepository(t)
	svc3 := NewAPIService(problems2, submissions, sessions, reviews, registry2, nil)
	problems2.EXPECT().UpsertFromManifest(ctx, manifest).Return(nil)
	registry2.EXPECT().Save(ctx, "v1", mock.Anything, []domain.ManifestRef(nil)).Return(errors.New("save"))
	_, err = svc3.SyncRegistry(ctx, "v1", time.Time{}, []domain.ProblemManifest{manifest}, nil)
	require.Error(t, err)
}

func TestAPIServiceSyncRegistrySuccess(t *testing.T) {
	problems := outmocks.NewMockProblemRepository(t)
	registry := outmocks.NewMockRegistryRepository(t)
	svc := NewAPIService(problems, nil, nil, nil, registry, nil)

	ctx := context.Background()
	updatedAt := time.Date(2026, 3, 31, 10, 0, 0, 0, time.UTC)
	problemsList := []domain.ProblemManifest{{Slug: "a"}, {Slug: "b"}}
	refs := []domain.ManifestRef{{Name: "main", Path: "providers/main.json"}}
	for _, problem := range problemsList {
		problems.EXPECT().UpsertFromManifest(ctx, problem).Return(nil)
	}
	registry.EXPECT().Save(ctx, "v1", updatedAt, refs).Return(nil)

	synced, err := svc.SyncRegistry(ctx, "v1", updatedAt, problemsList, refs)
	require.NoError(t, err)
	require.Equal(t, 2, synced)
}
