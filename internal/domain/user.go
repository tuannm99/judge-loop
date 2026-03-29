package domain

import (
	"time"

	"github.com/google/uuid"
)

// User is the primary account entity.
type User struct {
	ID        uuid.UUID
	Username  string
	Email     string
	CreatedAt time.Time
	UpdatedAt time.Time
}

// DifficultyMix defines the target percentage split across difficulty levels.
// Values should sum to 100.
type DifficultyMix struct {
	EasyPct   int // percentage of easy problems
	MediumPct int // percentage of medium problems
	HardPct   int // percentage of hard problems
}

// UserTrainingProfile stores a user's training goals and known weaknesses.
type UserTrainingProfile struct {
	ID             uuid.UUID
	UserID         uuid.UUID
	Goals          []string      // e.g. ["faang-interview", "competitive"]
	MinutesPerDay  int           // target daily practice time
	DifficultyMix  DifficultyMix // target difficulty distribution
	WeakPatterns   []string      // patterns the user struggles with
	FocusPatterns  []string      // patterns to prioritize right now
	UpdatedAt      time.Time
}

// TrainingContract defines the user's daily and weekly commitments.
type TrainingContract struct {
	ID             uuid.UUID
	UserID         uuid.UUID
	DailyProblems  int       // required solves per day
	WeeklyProblems int       // weekly target
	FocusTime      int       // minutes per session
	ReviewEnabled  bool      // spaced repetition on/off
	ActiveFrom     time.Time // when this contract took effect
	CreatedAt      time.Time
}
