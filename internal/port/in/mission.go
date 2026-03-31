package in

import (
	"context"

	"github.com/google/uuid"
	"github.com/tuannm99/judge-loop/internal/domain"
)

type MissionService interface {
	GetDailyMission(ctx context.Context, userID uuid.UUID) (*domain.DailyMission, error)
}
