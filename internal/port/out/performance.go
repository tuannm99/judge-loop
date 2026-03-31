package out

import (
	"context"

	"github.com/google/uuid"
)

type PerformanceRepository interface {
	GetPatternScores(ctx context.Context, userID uuid.UUID) (map[string]float64, error)
}
