package domain

import (
	"time"

	"github.com/google/uuid"
)

// Language is a supported submission language.
type Language string

const (
	LanguagePython Language = "python"
	LanguageGo     Language = "go"
)

// SubmissionStatus represents the lifecycle state of a submission.
type SubmissionStatus string

const (
	StatusPending            SubmissionStatus = "pending"
	StatusRunning            SubmissionStatus = "running"
	StatusAccepted           SubmissionStatus = "accepted"
	StatusWrongAnswer        SubmissionStatus = "wrong_answer"
	StatusCompileError       SubmissionStatus = "compile_error"
	StatusRuntimeError       SubmissionStatus = "runtime_error"
	StatusTimeLimitExceeded  SubmissionStatus = "time_limit_exceeded"
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
	ID           uuid.UUID
	UserID       uuid.UUID
	ProblemID    uuid.UUID
	SessionID    *uuid.UUID // optional: links to a TimerSession
	Language     Language
	Code         string
	Status       SubmissionStatus
	Verdict      Verdict
	PassedCases  int
	TotalCases   int
	RuntimeMS    int64  // execution time in milliseconds
	ErrorMessage string // populated on CE/RE
	SubmittedAt  time.Time
	EvaluatedAt  *time.Time
}
