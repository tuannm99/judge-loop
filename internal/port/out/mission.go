package out

import (
	"context"

	"github.com/google/uuid"
	"github.com/tuannm99/judge-loop/internal/domain"
)

type MissionRepository interface {
	GetToday(ctx context.Context, userID uuid.UUID) (*domain.DailyMission, error)
	Save(ctx context.Context, mission domain.DailyMission) error
}
