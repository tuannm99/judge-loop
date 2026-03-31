package in

import (
	"context"

	"github.com/google/uuid"
	"github.com/tuannm99/judge-loop/internal/domain"
)

type TimerService interface {
	StartTimer(ctx context.Context, userID uuid.UUID, problemID *uuid.UUID) (*domain.TimerSession, error)
	StopTimer(ctx context.Context, userID uuid.UUID) (*domain.TimerSession, error)
	CurrentTimer(ctx context.Context, userID uuid.UUID) (*domain.TimerSession, error)
}
