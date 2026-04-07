package in

import (
	"context"

	"github.com/google/uuid"
	out "github.com/tuannm99/judge-loop/internal/port/out"
)

const GoalTotal = 3000

type ProgressService interface {
	GetProgressToday(ctx context.Context, userID uuid.UUID) (ProgressToday, error)
	GetStreak(ctx context.Context, userID uuid.UUID) (out.StreakInfo, error)
	GetGoalProgress(ctx context.Context, userID uuid.UUID) (GoalProgress, error)
}

type ProgressToday struct {
	Date             string
	Solved           int
	Attempted        int
	TimeSpentMinutes int
	Streak           int
}

type GoalProgress struct {
	Solved    int     `json:"solved"`
	Total     int     `json:"total"`
	Remaining int     `json:"remaining"`
	Percent   float64 `json:"percent"`
}
