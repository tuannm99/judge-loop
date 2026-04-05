package domain

import (
	"time"

	"github.com/google/uuid"
)

// Difficulty represents the difficulty level of a problem.
type Difficulty string

const (
	DifficultyEasy   Difficulty = "easy"
	DifficultyMedium Difficulty = "medium"
	DifficultyHard   Difficulty = "hard"
)

// Provider is the source platform for a problem.
type Provider string

const (
	ProviderLeetCode   Provider = "leetcode"
	ProviderNeetCode   Provider = "neetcode"
	ProviderHackerRank Provider = "hackerrank"
)

// Problem is a practice problem stored in the local bank.
// Full problem statements are NOT stored — only metadata.
type Problem struct {
	ID            uuid.UUID         `json:"id"`
	Slug          string            `json:"slug"`
	Title         string            `json:"title"`
	Difficulty    Difficulty        `json:"difficulty"`
	Tags          []string          `json:"tags"`         // data structure / algorithm tags
	PatternTags   []string          `json:"pattern_tags"` // pattern tags (e.g. sliding-window)
	Provider      Provider          `json:"provider"`
	ExternalID    string            `json:"external_id"` // provider's own ID
	SourceURL     string            `json:"source_url"`
	EstimatedTime int               `json:"estimated_time"` // minutes
	StarterCode   map[string]string `json:"starter_code"`
	CreatedAt     time.Time         `json:"created_at"`
	UpdatedAt     time.Time         `json:"updated_at"`
}

type ProblemLabel struct {
	ID        uuid.UUID `json:"id"`
	Kind      string    `json:"kind"`
	Slug      string    `json:"slug"`
	Name      string    `json:"name"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// TestCase stores input/output for judge evaluation.
type TestCase struct {
	ID        uuid.UUID
	ProblemID uuid.UUID
	Input     string
	Expected  string
	IsHidden  bool // hidden test cases are not shown to the user
	OrderIdx  int  // display order
}

// ProblemPerformance tracks a user's history on a specific problem.
type ProblemPerformance struct {
	ID              uuid.UUID
	UserID          uuid.UUID
	ProblemID       uuid.UUID
	FirstSolveTime  *float64 // minutes, nil if not yet solved
	BestSolveTime   *float64 // personal best
	LatestSolveTime *float64 // most recent solve
	Attempts        int      // total attempts
	Accepted        bool     // ever accepted
	Complexity      string   // self-reported: O(n), O(n log n), etc.
	Confidence      int      // self-reported: 1 (shaky) to 5 (solid)
	LastAttemptAt   time.Time
	UpdatedAt       time.Time
}

// ProblemBankItem links a problem to the user's local problem bank.
// It tracks whether the problem has been downloaded/imported.
type ProblemBankItem struct {
	ID         uuid.UUID
	UserID     uuid.UUID
	ProblemID  uuid.UUID
	Imported   bool // manifest pulled locally
	InProgress bool // user is actively working on it
	Pinned     bool // user pinned it for later
	AddedAt    time.Time
}
