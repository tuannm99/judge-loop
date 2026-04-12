// Package judge evaluates submission results against expected test case outputs.
package judge

import (
	"fmt"
	"strings"

	"github.com/tuannm99/judge-loop/internal/domain"
)

type RunResult struct {
	Output    string
	Stderr    string
	ExitCode  int
	TimedOut  bool
	RuntimeMS int64
}

// RunFn executes code against a single test case input and returns the result.
type RunFn func(input string) (RunResult, error)

// Evaluate runs all test cases via runFn and returns the aggregate verdict.
// If there are no test cases, it returns Accepted (can't judge without test data).
func Evaluate(cases []domain.TestCase, runFn RunFn) (
	status domain.SubmissionStatus,
	verdict domain.Verdict,
	passed, total int,
	maxRuntimeMS int64,
	errMsg string,
) {
	total = len(cases)

	if total == 0 {
		// No test cases stored — treat as accepted (registry data comes in Milestone 8).
		return domain.StatusAccepted, domain.VerdictAccepted, 0, 0, 0, ""
	}

	for i, tc := range cases {
		result, err := runFn(tc.Input)
		if err != nil {
			return domain.StatusRuntimeError, domain.VerdictRuntimeError,
				passed, total, maxRuntimeMS, err.Error()
		}
		if result.TimedOut {
			return domain.StatusTimeLimitExceeded, domain.VerdictTimeLimitExceeded,
				passed, total, maxRuntimeMS, ""
		}
		if result.RuntimeMS > maxRuntimeMS {
			maxRuntimeMS = result.RuntimeMS
		}
		if result.ExitCode != 0 {
			if isCompileError(result.Stderr) {
				return domain.StatusCompileError, domain.VerdictCompileError,
					passed, total, maxRuntimeMS, result.Stderr
			}
			return domain.StatusRuntimeError, domain.VerdictRuntimeError,
				passed, total, maxRuntimeMS, result.Stderr
		}

		got := strings.TrimSpace(result.Output)
		exp := strings.TrimSpace(tc.Expected)
		if got == exp {
			passed++
		} else {
			msg := fmt.Sprintf("case %d: expected %q, got %q", i+1, exp, got)
			return domain.StatusWrongAnswer, domain.VerdictWrongAnswer,
				passed, total, maxRuntimeMS, msg
		}
	}

	return domain.StatusAccepted, domain.VerdictAccepted, passed, total, maxRuntimeMS, ""
}

// isCompileError heuristically checks if stderr looks like a compile/syntax error
// rather than a runtime error.
func isCompileError(stderr string) bool {
	lower := strings.ToLower(stderr)
	return strings.Contains(lower, "syntaxerror") ||
		strings.Contains(lower, "indentationerror") ||
		strings.Contains(lower, "nameerror: name") ||
		strings.Contains(lower, "go build") ||
		strings.Contains(lower, "syntax error") ||
		strings.Contains(lower, "undefined:") ||
		strings.Contains(lower, "could not be parsed") ||
		strings.Contains(lower, "type checking failed") ||
		strings.Contains(lower, "error[")
}
