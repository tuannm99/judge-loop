package application

import (
	"context"

	"github.com/google/uuid"
	"github.com/tuannm99/judge-loop/internal/domain"
	inport "github.com/tuannm99/judge-loop/internal/port/in"
	outport "github.com/tuannm99/judge-loop/internal/port/out"
)

type TimerService struct {
	sessions outport.SessionRepository
}

var _ inport.TimerService = (*TimerService)(nil)

func NewTimerService(sessions outport.SessionRepository) *TimerService {
	return &TimerService{sessions: sessions}
}

func (s *TimerService) StartTimer(ctx context.Context, userID uuid.UUID, problemID *uuid.UUID) (*domain.TimerSession, error) {
	return s.sessions.StartTimer(ctx, userID, problemID)
}

func (s *TimerService) StopTimer(ctx context.Context, userID uuid.UUID) (*domain.TimerSession, error) {
	return s.sessions.StopTimer(ctx, userID)
}

func (s *TimerService) CurrentTimer(ctx context.Context, userID uuid.UUID) (*domain.TimerSession, error) {
	return s.sessions.ActiveTimer(ctx, userID)
}
