package out

import (
	"time"

	"github.com/google/uuid"
	"github.com/tuannm99/judge-loop/internal/domain"
)

type ProblemFilter struct {
	Difficulty *domain.Difficulty
	Tags       []string
	Patterns   []string
	Provider   *domain.Provider
	Limit      int
	Offset     int
}

type StreakInfo struct {
	Current       int
	Longest       int
	LastPracticed *time.Time
}

type DueReview struct {
	ProblemID   uuid.UUID
	Slug        string
	Title       string
	LastSolved  *time.Time
	DaysOverdue int
}

type RegistryVersion struct {
	Version   string
	UpdatedAt time.Time
	SyncedAt  time.Time
}

type EvaluateSubmissionJob struct {
	SubmissionID string
	UserID       string
}

type RunRequest struct {
	Language string
	Code     string
	Input    string
}
