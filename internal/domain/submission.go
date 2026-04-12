package domain

import (
	"errors"
	"strings"
	"time"

	"github.com/google/uuid"
)

// Language is a supported submission language.
type Language string

const (
	LanguagePython     Language = "python"
	LanguageGo         Language = "go"
	LanguageJavascript Language = "javascript"
	LanguageTypescript Language = "typescript"
	LanguageRust       Language = "rust"
)

var ErrUnsupportedSubmissionLanguage = errors.New("unsupported submission language")

var supportedSubmissionLanguages = []Language{
	LanguagePython,
	LanguageGo,
	LanguageJavascript,
	LanguageTypescript,
	LanguageRust,
}

func NormalizeSubmissionLanguage(raw string) Language {
	return Language(strings.ToLower(strings.TrimSpace(raw)))
}

func IsSupportedSubmissionLanguage(raw string) bool {
	normalized := NormalizeSubmissionLanguage(raw)
	for _, language := range supportedSubmissionLanguages {
		if normalized == language {
			return true
		}
	}
	return false
}

// SubmissionStatus represents the lifecycle state of a submission.
type SubmissionStatus string

const (
	StatusPending           SubmissionStatus = "pending"
	StatusRunning           SubmissionStatus = "running"
	StatusAccepted          SubmissionStatus = "accepted"
	StatusWrongAnswer       SubmissionStatus = "wrong_answer"
	StatusCompileError      SubmissionStatus = "compile_error"
	StatusRuntimeError      SubmissionStatus = "runtime_error"
	StatusTimeLimitExceeded SubmissionStatus = "time_limit_exceeded"
)

// Verdict is the human-readable verdict label.
type Verdict string

const (
	VerdictAccepted          Verdict = "Accepted"
	VerdictWrongAnswer       Verdict = "Wrong Answer"
	VerdictCompileError      Verdict = "Compile Error"
	VerdictRuntimeError      Verdict = "Runtime Error"
	VerdictTimeLimitExceeded Verdict = "Time Limit Exceeded"
)

// Submission is a user's code submission for a problem.
type Submission struct {
	ID           uuid.UUID        `json:"id"`
	UserID       uuid.UUID        `json:"user_id"`
	ProblemID    uuid.UUID        `json:"problem_id"`
	SessionID    *uuid.UUID       `json:"session_id,omitempty"`
	Language     Language         `json:"language"`
	Code         string           `json:"code"`
	Status       SubmissionStatus `json:"status"`
	Verdict      Verdict          `json:"verdict"`
	PassedCases  int              `json:"passed_cases"`
	TotalCases   int              `json:"total_cases"`
	RuntimeMS    int64            `json:"runtime_ms"`
	ErrorMessage string           `json:"error_message"`
	SubmittedAt  time.Time        `json:"submitted_at"`
	EvaluatedAt  *time.Time       `json:"evaluated_at,omitempty"`
}
