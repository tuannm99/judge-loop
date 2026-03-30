package domain

import (
	"time"

	"github.com/google/uuid"
)

// MissionTask is a single problem task within a daily mission.
type MissionTask struct {
	ProblemID  uuid.UUID  `json:"problem_id"`
	Slug       string     `json:"slug"`
	Title      string     `json:"title"`
	Difficulty Difficulty `json:"difficulty"`
	Reason     string     `json:"reason"`   // why selected, e.g. "weak pattern: sliding-window"
	Priority   int        `json:"priority"` // higher = more important
}

// DailyMission is generated once per day per user.
type DailyMission struct {
	ID            uuid.UUID     `json:"id"`
	UserID        uuid.UUID     `json:"user_id"`
	Date          time.Time     `json:"date"`
	RequiredTasks []MissionTask `json:"required_tasks"` // must complete for daily goal
	OptionalTasks []MissionTask `json:"optional_tasks"` // bonus problems
	ReviewTasks   []MissionTask `json:"review_tasks"`   // spaced repetition reviews
	GeneratedAt   time.Time     `json:"generated_at"`
}

// ReviewSchedule tracks when a problem should next be reviewed.
// Uses simple interval-based spaced repetition.
type ReviewSchedule struct {
	ID           uuid.UUID `json:"id"`
	UserID       uuid.UUID `json:"user_id"`
	ProblemID    uuid.UUID `json:"problem_id"`
	NextReviewAt time.Time `json:"next_review_at"`
	IntervalDays int       `json:"interval_days"`
	ReviewCount  int       `json:"review_count"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}
