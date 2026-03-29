package domain

import (
	"time"

	"github.com/google/uuid"
)

// PerformanceSnapshot is a periodic snapshot of user-wide performance.
// Taken weekly to enable self-comparison over time.
type PerformanceSnapshot struct {
	ID             uuid.UUID
	UserID         uuid.UUID
	SnapshotDate   time.Time
	AvgSolveTime   float64            // minutes, across accepted submissions in the period
	TotalAttempts  int
	AcceptedCount  int
	HintUsageRate  float64            // 0.0 to 1.0 (reserved for future hint tracking)
	PatternScores  map[string]float64 // pattern tag → score (accepted / attempted)
	CreatedAt      time.Time
}
