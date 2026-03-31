package out

import (
	"context"

	"github.com/google/uuid"
	"github.com/tuannm99/judge-loop/internal/domain"
)

type SessionRepository interface {
	GetOrCreateToday(ctx context.Context, userID uuid.UUID) (*domain.DailySession, error)
	RecordSubmission(ctx context.Context, userID uuid.UUID, accepted bool) error
	StartTimer(ctx context.Context, userID uuid.UUID, problemID *uuid.UUID) (*domain.TimerSession, error)
	StopTimer(ctx context.Context, userID uuid.UUID) (*domain.TimerSession, error)
	ActiveTimer(ctx context.Context, userID uuid.UUID) (*domain.TimerSession, error)
	GetStreak(ctx context.Context, userID uuid.UUID) (StreakInfo, error)
}
