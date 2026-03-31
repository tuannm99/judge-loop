package application

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"github.com/tuannm99/judge-loop/internal/domain"
	outport "github.com/tuannm99/judge-loop/internal/port/out"
	outmocks "github.com/tuannm99/judge-loop/internal/port/out/mocks"
)

func TestMissionServiceReturnsCachedMission(t *testing.T) {
	userID := uuid.New()
	cached := &domain.DailyMission{ID: uuid.New(), UserID: userID}

	missions := outmocks.NewMockMissionRepository(t)
	missions.EXPECT().GetToday(mock.Anything, userID).Return(cached, nil).Once()

	svc := NewMissionService(missions, nil, nil, nil)

	got, err := svc.GetDailyMission(context.Background(), userID)
	require.NoError(t, err)
	require.Same(t, cached, got)
}

func TestMissionServiceGeneratesAndSavesMission(t *testing.T) {
	userID := uuid.New()
	reviewID := uuid.New()
	weakID := uuid.New()
	otherID := uuid.New()

	missions := outmocks.NewMockMissionRepository(t)
	performance := outmocks.NewMockPerformanceRepository(t)
	problems := outmocks.NewMockProblemRepository(t)
	reviews := outmocks.NewMockReviewRepository(t)

	missions.EXPECT().GetToday(mock.Anything, userID).Return(nil, nil).Once()
	performance.EXPECT().GetPatternScores(mock.Anything, userID).Return(map[string]float64{"dp": 0.2}, nil).Once()
	problems.EXPECT().GetUnsolved(mock.Anything, userID, defaultMissionProblemLimit).Return([]domain.Problem{
		{ID: weakID, Slug: "weak", Title: "Weak", Difficulty: domain.DifficultyEasy, PatternTags: []string{"dp"}},
		{ID: otherID, Slug: "other", Title: "Other", Difficulty: domain.DifficultyMedium, PatternTags: []string{"array"}},
	}, nil).Once()
	reviews.EXPECT().GetDue(mock.Anything, userID).Return([]outport.DueReview{
		{ProblemID: reviewID, Slug: "review", Title: "Review", DaysOverdue: 1},
	}, nil).Once()

	var saved domain.DailyMission
	missions.EXPECT().Save(mock.Anything, mock.MatchedBy(func(mission domain.DailyMission) bool {
		saved = mission
		return mission.UserID == userID &&
			len(mission.ReviewTasks) == 1 &&
			mission.ReviewTasks[0].ProblemID == reviewID &&
			len(mission.RequiredTasks) == 2 &&
			mission.RequiredTasks[0].ProblemID == weakID
	})).Return(nil).Once()

	svc := NewMissionService(missions, performance, problems, reviews)

	got, err := svc.GetDailyMission(context.Background(), userID)
	require.NoError(t, err)
	require.NotNil(t, got)
	require.Equal(t, userID, got.UserID)
	require.Equal(t, got.ID, saved.ID)
	require.Len(t, got.ReviewTasks, 1)
	require.Equal(t, reviewID, got.ReviewTasks[0].ProblemID)
	require.Len(t, got.RequiredTasks, 2)
	require.Equal(t, weakID, got.RequiredTasks[0].ProblemID)
}

func TestMissionServiceReturnsDependencyErrors(t *testing.T) {
	userID := uuid.New()

	missions := outmocks.NewMockMissionRepository(t)
	missions.EXPECT().GetToday(mock.Anything, userID).Return(nil, errors.New("mission")).Once()
	_, err := NewMissionService(missions, nil, nil, nil).GetDailyMission(context.Background(), userID)
	require.ErrorContains(t, err, "mission")

	missions = outmocks.NewMockMissionRepository(t)
	performance := outmocks.NewMockPerformanceRepository(t)
	missions.EXPECT().GetToday(mock.Anything, userID).Return(nil, nil).Once()
	performance.EXPECT().GetPatternScores(mock.Anything, userID).Return(nil, errors.New("scores")).Once()
	_, err = NewMissionService(missions, performance, nil, nil).GetDailyMission(context.Background(), userID)
	require.ErrorContains(t, err, "scores")

	missions = outmocks.NewMockMissionRepository(t)
	problems := outmocks.NewMockProblemRepository(t)
	missions.EXPECT().GetToday(mock.Anything, userID).Return(nil, nil).Once()
	problems.EXPECT().GetUnsolved(mock.Anything, userID, defaultMissionProblemLimit).Return(nil, errors.New("unsolved")).Once()
	_, err = NewMissionService(missions, nil, problems, nil).GetDailyMission(context.Background(), userID)
	require.ErrorContains(t, err, "unsolved")

	missions = outmocks.NewMockMissionRepository(t)
	reviews := outmocks.NewMockReviewRepository(t)
	missions.EXPECT().GetToday(mock.Anything, userID).Return(nil, nil).Once()
	reviews.EXPECT().GetDue(mock.Anything, userID).Return(nil, errors.New("reviews")).Once()
	_, err = NewMissionService(missions, nil, nil, reviews).GetDailyMission(context.Background(), userID)
	require.ErrorContains(t, err, "reviews")

	missions = outmocks.NewMockMissionRepository(t)
	missions.EXPECT().GetToday(mock.Anything, userID).Return(nil, nil).Once()
	missions.EXPECT().Save(mock.Anything, mock.Anything).Return(errors.New("save")).Once()
	_, err = NewMissionService(missions, nil, nil, nil).GetDailyMission(context.Background(), userID)
	require.ErrorContains(t, err, "save")
}
