// Package localtimer provides an in-memory timer for tracking active practice sessions.
// State is lost on process restart.
package localtimer

import (
	"sync"
	"time"

	"github.com/google/uuid"
)

// Entry represents a single timed practice session.
type Entry struct {
	ID        uuid.UUID
	ServerID  *uuid.UUID
	ProblemID *uuid.UUID
	StartedAt time.Time
}

// LocalTimer holds the state of the currently active timer.
// All methods are safe for concurrent use.
type LocalTimer struct {
	mu     sync.Mutex
	active *Entry
}

// New creates a new LocalTimer with no active session.
func New() *LocalTimer {
	return &LocalTimer{}
}

// Start begins a new timer session, stopping any existing one first.
// Returns the new entry.
func (t *LocalTimer) Start(problemID *uuid.UUID) Entry {
	t.mu.Lock()
	defer t.mu.Unlock()

	e := Entry{
		ID:        uuid.New(),
		ProblemID: problemID,
		StartedAt: time.Now(),
	}
	t.active = &e
	return e
}

// SetServerID records the authoritative api-server timer session ID for the
// active local entry. It is a no-op if the local entry is no longer active.
func (t *LocalTimer) SetServerID(localID, serverID uuid.UUID) {
	t.mu.Lock()
	defer t.mu.Unlock()

	if t.active == nil || t.active.ID != localID {
		return
	}
	t.active.ServerID = &serverID
}

// Stop ends the active timer session.
// Returns the stopped entry and true if a timer was running,
// or a zero Entry and false if nothing was active.
func (t *LocalTimer) Stop() (Entry, bool) {
	t.mu.Lock()
	defer t.mu.Unlock()

	if t.active == nil {
		return Entry{}, false
	}
	e := *t.active
	t.active = nil
	return e, true
}

// Active returns a copy of the active entry, or nil if no timer is running.
func (t *LocalTimer) Active() *Entry {
	t.mu.Lock()
	defer t.mu.Unlock()

	if t.active == nil {
		return nil
	}
	cp := *t.active
	return &cp
}

// ElapsedSecs returns how many seconds have elapsed since the timer started.
// Returns 0 if no timer is active.
func (t *LocalTimer) ElapsedSecs() int {
	t.mu.Lock()
	defer t.mu.Unlock()

	if t.active == nil {
		return 0
	}
	return int(time.Since(t.active.StartedAt).Seconds())
}
