package domain

import (
	"encoding/json"
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
// It stores metadata and may include an optional author-written Markdown description.
type Problem struct {
	ID                  uuid.UUID         `json:"id"`
	Slug                string            `json:"slug"`
	Title               string            `json:"title"`
	Difficulty          Difficulty        `json:"difficulty"`
	Tags                []string          `json:"tags"` // combined taxonomy tags
	Provider            Provider          `json:"provider"`
	ExternalID          string            `json:"external_id"` // provider's own ID
	SourceURL           string            `json:"source_url"`
	EstimatedTime       int               `json:"estimated_time"` // minutes
	DescriptionMarkdown string            `json:"description_markdown"`
	StarterCode         map[string]string `json:"starter_code"`
	ExecutionSpec       ExecutionSpec     `json:"execution_spec"`
	JudgeReady          bool              `json:"judge_ready"`
	CreatedAt           time.Time         `json:"created_at"`
	UpdatedAt           time.Time         `json:"updated_at"`
}

type ExecutionMode string

const (
	ExecutionModeStdin       ExecutionMode = "stdin"
	ExecutionModeFunction    ExecutionMode = "function"
	ExecutionModeClass       ExecutionMode = "class"
	ExecutionModeInPlace     ExecutionMode = "in_place"
	ExecutionModeInteractive ExecutionMode = "interactive"
	ExecutionModeCustom      ExecutionMode = "custom"
)

type ExecutionSpec struct {
	Mode               ExecutionMode                    `json:"mode"`
	Entrypoint         string                           `json:"entrypoint,omitempty"`
	ClassName          string                           `json:"class_name,omitempty"`
	Signature          ExecutionSignature               `json:"signature,omitempty"`
	Constructor        ExecutionSignature               `json:"constructor,omitempty"`
	Methods            map[string]MethodSpec            `json:"methods,omitempty"`
	Output             ExecutionOutput                  `json:"output,omitempty"`
	Bindings           map[string]ExecutionLanguageBind `json:"bindings,omitempty"`
	SupportedLanguages []Language                       `json:"supported_languages,omitempty"`
	Comparator         ComparatorSpec                   `json:"comparator,omitempty"`
	TimeoutMS          int                              `json:"timeout_ms,omitempty"`
	MemoryMB           int                              `json:"memory_mb,omitempty"`
}

type ExecutionSignature struct {
	Params  []ExecutionParam `json:"params,omitempty"`
	Returns string           `json:"returns,omitempty"`
}

type ExecutionParam struct {
	Name string `json:"name"`
	Type string `json:"type"`
}

type MethodSpec struct {
	Params   []ExecutionParam  `json:"params,omitempty"`
	Returns  string            `json:"returns,omitempty"`
	Bindings map[string]string `json:"bindings,omitempty"`
}

type ExecutionOutput struct {
	Source     string `json:"source,omitempty"`
	ParamIndex int    `json:"param_index,omitempty"`
}

type ExecutionLanguageBind struct {
	Entrypoint  string `json:"entrypoint,omitempty"`
	ClassName   string `json:"class_name,omitempty"`
	Constructor string `json:"constructor,omitempty"`
}

type ComparatorSpec struct {
	Kind    string  `json:"kind,omitempty"`
	Epsilon float64 `json:"epsilon,omitempty"`
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
	ID           uuid.UUID       `json:"id"`
	ProblemID    uuid.UUID       `json:"problem_id"`
	Name         string          `json:"name"`
	Input        string          `json:"input"`
	Expected     string          `json:"expected"`
	InputJSON    json.RawMessage `json:"input_json,omitempty"`
	ExpectedJSON json.RawMessage `json:"expected_json,omitempty"`
	Metadata     json.RawMessage `json:"metadata,omitempty"`
	IsHidden     bool            `json:"is_hidden"` // hidden test cases are not shown to the user
	OrderIdx     int             `json:"order_idx"` // display order
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
