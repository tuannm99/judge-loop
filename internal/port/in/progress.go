package in

import (
	"context"

	"github.com/google/uuid"
	out "github.com/tuannm99/judge-loop/internal/port/out"
)

type ProgressService interface {
	GetProgressToday(ctx context.Context, userID uuid.UUID) (ProgressToday, error)
	GetStreak(ctx context.Context, userID uuid.UUID) (out.StreakInfo, error)
}

type ProgressToday struct {
	Date             string
	Solved           int
	Attempted        int
	TimeSpentMinutes int
	Streak           int
}
