package domain

import (
	"time"

	"github.com/google/uuid"
)

// DailySession tracks whether the user practiced on a given day.
type DailySession struct {
	ID             uuid.UUID
	UserID         uuid.UUID
	Date           time.Time // date only (UTC), truncated to day
	SolvedCount    int       // accepted submissions that day
	AttemptedCount int
	TimeSpentSecs  int // total seconds in timer sessions
	CreatedAt      time.Time
	UpdatedAt      time.Time
}

// TimerSession is a timed practice period started by the user.
type TimerSession struct {
	ID          uuid.UUID
	UserID      uuid.UUID
	ProblemID   *uuid.UUID // optional: problem being worked on
	StartedAt   time.Time
	EndedAt     *time.Time // nil while active
	ElapsedSecs int        // computed on stop
}

// ActivityEvent records discrete user actions for analytics.
type ActivityEvent struct {
	ID        uuid.UUID
	UserID    uuid.UUID
	EventType ActivityEventType
	ProblemID *uuid.UUID
	SessionID *uuid.UUID
	Payload   map[string]any // event-specific data
	OccuredAt time.Time
}

// ActivityEventType enumerates the kinds of activity events.
type ActivityEventType string

const (
	EventTimerStart    ActivityEventType = "timer_start"
	EventTimerStop     ActivityEventType = "timer_stop"
	EventSubmit        ActivityEventType = "submit"
	EventVerdictResult ActivityEventType = "verdict_result"
	EventSyncRegistry  ActivityEventType = "sync_registry"
)
