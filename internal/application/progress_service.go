package application

import (
	"context"

	"github.com/google/uuid"
	inport "github.com/tuannm99/judge-loop/internal/port/in"
	outport "github.com/tuannm99/judge-loop/internal/port/out"
)

type ProgressService struct {
	sessions outport.SessionRepository
}

var _ inport.ProgressService = (*ProgressService)(nil)

func NewProgressService(sessions outport.SessionRepository) *ProgressService {
	return &ProgressService{sessions: sessions}
}

func (s *ProgressService) GetProgressToday(ctx context.Context, userID uuid.UUID) (inport.ProgressToday, error) {
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

func (s *ProgressService) GetStreak(ctx context.Context, userID uuid.UUID) (outport.StreakInfo, error) {
	return s.sessions.GetStreak(ctx, userID)
}
