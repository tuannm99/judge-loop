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

func TestAllInportMocks(t *testing.T) {
	ctx := context.Background()
	userID := uuid.New()
	problemID := uuid.New()
	submissionID := uuid.New()
	sessionID := uuid.New()
	problem := &domain.Problem{ID: problemID, Slug: "two-sum"}
	submission := &domain.Submission{ID: submissionID}
	timer := &domain.TimerSession{ID: sessionID}
	progress := inport.ProgressToday{Date: "2026-04-01", Solved: 2, Attempted: 3, TimeSpentMinutes: 45, Streak: 7}
	streak := outport.StreakInfo{Current: 7, Longest: 9}
	reviews := []outport.DueReview{{ProblemID: problemID}}
	registryVersion := &outport.RegistryVersion{Version: "v1"}
	manifests := []domain.ManifestRef{{Name: "main", Path: "providers/main.json"}}
	problems := []domain.Problem{{ID: problemID}}
	submissions := []domain.Submission{{ID: submissionID}}
	mission := &domain.DailyMission{ID: uuid.New(), UserID: userID}

	problemSvc := NewMockProblemService(t)
	require.NotNil(t, problemSvc.EXPECT())
	filter := outport.ProblemFilter{Limit: 10}
	problemSvc.EXPECT().ListProblems(mock.Anything, filter).Run(func(context.Context, outport.ProblemFilter) {}).Return(problems, 1, nil)
	gotProblems, total, err := problemSvc.ListProblems(ctx, filter)
	require.NoError(t, err)
	require.Len(t, gotProblems, 1)
	require.Equal(t, 1, total)
	problemSvc.EXPECT().ListProblems(mock.Anything, filter).RunAndReturn(func(context.Context, outport.ProblemFilter) ([]domain.Problem, int, error) {
		return problems, 1, nil
	})
	_, _, err = problemSvc.ListProblems(ctx, filter)
	require.NoError(t, err)
	problemSvc.EXPECT().GetProblem(mock.Anything, "two-sum").Run(func(context.Context, string) {}).Return(problem, nil)
	gotProblem, err := problemSvc.GetProblem(ctx, "two-sum")
	require.NoError(t, err)
	require.Equal(t, problemID, gotProblem.ID)
	problemSvc.EXPECT().GetProblem(mock.Anything, "two-sum-2").RunAndReturn(func(context.Context, string) (*domain.Problem, error) {
		return problem, nil
	})
	_, err = problemSvc.GetProblem(ctx, "two-sum-2")
	require.NoError(t, err)
	problemSvc.EXPECT().SuggestProblem(mock.Anything, userID).Run(func(context.Context, uuid.UUID) {}).Return(problem, nil)
	_, err = problemSvc.SuggestProblem(ctx, userID)
	require.NoError(t, err)
	problemSvc.EXPECT().SuggestProblem(mock.Anything, userID).RunAndReturn(func(context.Context, uuid.UUID) (*domain.Problem, error) {
		return problem, nil
	})
	_, err = problemSvc.SuggestProblem(ctx, userID)
	require.NoError(t, err)

	submissionSvc := NewMockSubmissionService(t)
	require.NotNil(t, submissionSvc.EXPECT())
	submissionSvc.EXPECT().CreateSubmission(mock.Anything, userID, problemID, "go", "code", (*uuid.UUID)(nil)).
		Run(func(context.Context, uuid.UUID, uuid.UUID, string, string, *uuid.UUID) {}).
		Return(submission, nil)
	gotSubmission, err := submissionSvc.CreateSubmission(ctx, userID, problemID, "go", "code", nil)
	require.NoError(t, err)
	require.Equal(t, submissionID, gotSubmission.ID)
	submissionSvc.EXPECT().CreateSubmission(mock.Anything, userID, problemID, "go", "code", (*uuid.UUID)(nil)).
		RunAndReturn(func(context.Context, uuid.UUID, uuid.UUID, string, string, *uuid.UUID) (*domain.Submission, error) {
			return submission, nil
		})
	_, err = submissionSvc.CreateSubmission(ctx, userID, problemID, "go", "code", nil)
	require.NoError(t, err)
	submissionSvc.EXPECT().GetSubmission(mock.Anything, submissionID).Run(func(context.Context, uuid.UUID) {}).Return(submission, nil)
	_, err = submissionSvc.GetSubmission(ctx, submissionID)
	require.NoError(t, err)
	submissionSvc.EXPECT().GetSubmission(mock.Anything, submissionID).RunAndReturn(func(context.Context, uuid.UUID) (*domain.Submission, error) {
		return submission, nil
	})
	_, err = submissionSvc.GetSubmission(ctx, submissionID)
	require.NoError(t, err)
	submissionSvc.EXPECT().ListSubmissions(mock.Anything, userID, (*uuid.UUID)(nil), 10, 0).
		Run(func(context.Context, uuid.UUID, *uuid.UUID, int, int) {}).
		Return(submissions, nil)
	_, err = submissionSvc.ListSubmissions(ctx, userID, nil, 10, 0)
	require.NoError(t, err)
	submissionSvc.EXPECT().ListSubmissions(mock.Anything, userID, (*uuid.UUID)(nil), 10, 0).
		RunAndReturn(func(context.Context, uuid.UUID, *uuid.UUID, int, int) ([]domain.Submission, error) {
			return submissions, nil
		})
	_, err = submissionSvc.ListSubmissions(ctx, userID, nil, 10, 0)
	require.NoError(t, err)

	progressSvc := NewMockProgressService(t)
	require.NotNil(t, progressSvc.EXPECT())
	progressSvc.EXPECT().GetProgressToday(mock.Anything, userID).Run(func(context.Context, uuid.UUID) {}).Return(progress, nil)
	gotProgress, err := progressSvc.GetProgressToday(ctx, userID)
	require.NoError(t, err)
	require.Equal(t, 2, gotProgress.Solved)
	progressSvc.EXPECT().GetProgressToday(mock.Anything, userID).RunAndReturn(func(context.Context, uuid.UUID) (inport.ProgressToday, error) {
		return progress, nil
	})
	_, err = progressSvc.GetProgressToday(ctx, userID)
	require.NoError(t, err)
	progressSvc.EXPECT().GetStreak(mock.Anything, userID).Run(func(context.Context, uuid.UUID) {}).Return(streak, nil)
	_, err = progressSvc.GetStreak(ctx, userID)
	require.NoError(t, err)
	progressSvc.EXPECT().GetStreak(mock.Anything, userID).RunAndReturn(func(context.Context, uuid.UUID) (outport.StreakInfo, error) {
		return streak, nil
	})
	_, err = progressSvc.GetStreak(ctx, userID)
	require.NoError(t, err)

	timerSvc := NewMockTimerService(t)
	require.NotNil(t, timerSvc.EXPECT())
	timerSvc.EXPECT().StartTimer(mock.Anything, userID, (*uuid.UUID)(nil)).Run(func(context.Context, uuid.UUID, *uuid.UUID) {}).Return(timer, nil)
	_, err = timerSvc.StartTimer(ctx, userID, nil)
	require.NoError(t, err)
	timerSvc.EXPECT().StartTimer(mock.Anything, userID, (*uuid.UUID)(nil)).RunAndReturn(func(context.Context, uuid.UUID, *uuid.UUID) (*domain.TimerSession, error) {
		return timer, nil
	})
	_, err = timerSvc.StartTimer(ctx, userID, nil)
	require.NoError(t, err)
	timerSvc.EXPECT().StopTimer(mock.Anything, userID).Run(func(context.Context, uuid.UUID) {}).Return(timer, nil)
	_, err = timerSvc.StopTimer(ctx, userID)
	require.NoError(t, err)
	timerSvc.EXPECT().StopTimer(mock.Anything, userID).RunAndReturn(func(context.Context, uuid.UUID) (*domain.TimerSession, error) {
		return timer, nil
	})
	_, err = timerSvc.StopTimer(ctx, userID)
	require.NoError(t, err)
	timerSvc.EXPECT().CurrentTimer(mock.Anything, userID).Run(func(context.Context, uuid.UUID) {}).Return(timer, nil)
	_, err = timerSvc.CurrentTimer(ctx, userID)
	require.NoError(t, err)
	timerSvc.EXPECT().CurrentTimer(mock.Anything, userID).RunAndReturn(func(context.Context, uuid.UUID) (*domain.TimerSession, error) {
		return timer, nil
	})
	_, err = timerSvc.CurrentTimer(ctx, userID)
	require.NoError(t, err)

	reviewSvc := NewMockReviewService(t)
	require.NotNil(t, reviewSvc.EXPECT())
	reviewSvc.EXPECT().GetReviewsToday(mock.Anything, userID).Run(func(context.Context, uuid.UUID) {}).Return(reviews, nil)
	_, err = reviewSvc.GetReviewsToday(ctx, userID)
	require.NoError(t, err)
	reviewSvc.EXPECT().GetReviewsToday(mock.Anything, userID).RunAndReturn(func(context.Context, uuid.UUID) ([]outport.DueReview, error) {
		return reviews, nil
	})
	_, err = reviewSvc.GetReviewsToday(ctx, userID)
	require.NoError(t, err)

	registrySvc := NewMockRegistryService(t)
	require.NotNil(t, registrySvc.EXPECT())
	now := time.Now()
	problemManifests := []domain.ProblemManifest{{Slug: "two-sum"}}
	registrySvc.EXPECT().SyncRegistry(mock.Anything, "v1", now, problemManifests, manifests).
		Run(func(context.Context, string, time.Time, []domain.ProblemManifest, []domain.ManifestRef) {}).
		Return(1, nil)
	synced, err := registrySvc.SyncRegistry(ctx, "v1", now, problemManifests, manifests)
	require.NoError(t, err)
	require.Equal(t, 1, synced)
	registrySvc.EXPECT().SyncRegistry(mock.Anything, "v1", now, problemManifests, manifests).
		RunAndReturn(func(context.Context, string, time.Time, []domain.ProblemManifest, []domain.ManifestRef) (int, error) {
			return 1, nil
		})
	_, err = registrySvc.SyncRegistry(ctx, "v1", now, problemManifests, manifests)
	require.NoError(t, err)
	registrySvc.EXPECT().GetRegistryVersion(mock.Anything).Run(func(context.Context) {}).Return(registryVersion, nil)
	_, err = registrySvc.GetRegistryVersion(ctx)
	require.NoError(t, err)
	registrySvc.EXPECT().GetRegistryVersion(mock.Anything).RunAndReturn(func(context.Context) (*outport.RegistryVersion, error) {
		return registryVersion, nil
	})
	_, err = registrySvc.GetRegistryVersion(ctx)
	require.NoError(t, err)

	missionSvc := NewMockMissionService(t)
	require.NotNil(t, missionSvc.EXPECT())
	missionSvc.EXPECT().GetDailyMission(mock.Anything, userID).Run(func(context.Context, uuid.UUID) {}).Return(mission, nil)
	_, err = missionSvc.GetDailyMission(ctx, userID)
	require.NoError(t, err)
	missionSvc.EXPECT().GetDailyMission(mock.Anything, userID).RunAndReturn(func(context.Context, uuid.UUID) (*domain.DailyMission, error) {
		return mission, nil
	})
	_, err = missionSvc.GetDailyMission(ctx, userID)
	require.NoError(t, err)
}
