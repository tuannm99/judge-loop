package domain

import (
	"time"

	"github.com/google/uuid"
)

// MissionTask is a single problem task within a daily mission.
type MissionTask struct {
	ProblemID uuid.UUID
	Slug      string
	Title     string
	Difficulty Difficulty
	Reason    string // why this was selected (e.g. "weak pattern: sliding-window")
	Priority  int    // higher = more important
}

// DailyMission is generated once per day per user.
type DailyMission struct {
	ID            uuid.UUID
	UserID        uuid.UUID
	Date          time.Time      // UTC day
	RequiredTasks []MissionTask  // must complete for daily goal
	OptionalTasks []MissionTask  // bonus problems
	ReviewTasks   []MissionTask  // spaced repetition reviews
	GeneratedAt   time.Time
}

// ReviewSchedule tracks when a problem should next be reviewed.
// Uses simple interval-based spaced repetition.
type ReviewSchedule struct {
	ID            uuid.UUID
	UserID        uuid.UUID
	ProblemID     uuid.UUID
	NextReviewAt  time.Time
	IntervalDays  int    // current interval length
	ReviewCount   int    // how many times reviewed
	CreatedAt     time.Time
	UpdatedAt     time.Time
}
