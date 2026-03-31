package application

import (
	"context"

	"github.com/google/uuid"
	"github.com/tuannm99/judge-loop/internal/domain"
	inport "github.com/tuannm99/judge-loop/internal/port/in"
	outport "github.com/tuannm99/judge-loop/internal/port/out"
)

const defaultMissionProblemLimit = 10

type MissionService struct {
	missions    outport.MissionRepository
	performance outport.PerformanceRepository
	problems    outport.ProblemRepository
	reviews     outport.ReviewRepository
}

var _ inport.MissionService = (*MissionService)(nil)

func NewMissionService(
	missions outport.MissionRepository,
	performance outport.PerformanceRepository,
	problems outport.ProblemRepository,
	reviews outport.ReviewRepository,
) *MissionService {
	return &MissionService{
		missions:    missions,
		performance: performance,
		problems:    problems,
		reviews:     reviews,
	}
}

func (s *MissionService) GetDailyMission(ctx context.Context, userID uuid.UUID) (*domain.DailyMission, error) {
	if s.missions == nil {
		return nil, nil
	}

	current, err := s.missions.GetToday(ctx, userID)
	if err != nil {
		return nil, err
	}
	if current != nil {
		return current, nil
	}

	input := Input{}
	if s.performance != nil {
		scores, err := s.performance.GetPatternScores(ctx, userID)
		if err != nil {
			return nil, err
		}
		input.PatternScores = scores
	}
	if s.problems != nil {
		unsolved, err := s.problems.GetUnsolved(ctx, userID, defaultMissionProblemLimit)
		if err != nil {
			return nil, err
		}
		input.Unsolved = unsolved
	}
	if s.reviews != nil {
		due, err := s.reviews.GetDue(ctx, userID)
		if err != nil {
			return nil, err
		}
		input.DueReviews = due
	}

	mission := Generate(userID, input)
	if err := s.missions.Save(ctx, mission); err != nil {
		return nil, err
	}
	return &mission, nil
}
