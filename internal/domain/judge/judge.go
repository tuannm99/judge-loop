// Package judge evaluates submission results against expected test case outputs.
package judge

import (
	"encoding/json"
	"fmt"
	"math"
	"reflect"
	"sort"
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

// RunFn executes code against a single test case and returns the result.
type RunFn func(tc domain.TestCase) (RunResult, error)

// Evaluate runs all test cases via runFn and returns the aggregate verdict.
// If there are no test cases, it returns Accepted (can't judge without test data).
func Evaluate(cases []domain.TestCase, runFn RunFn) (
	status domain.SubmissionStatus,
	verdict domain.Verdict,
	passed, total int,
	maxRuntimeMS int64,
	errMsg string,
) {
	return EvaluateWithSpec(cases, domain.ExecutionSpec{}, runFn)
}

func EvaluateWithSpec(cases []domain.TestCase, spec domain.ExecutionSpec, runFn RunFn) (
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
		result, err := runFn(tc)
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

		got := normalizeOutput(result.Output)
		exp := expectedOutput(tc)
		if outputsEqualWithComparator(got, exp, spec.Comparator) {
			passed++
		} else {
			msg := fmt.Sprintf("case %d: expected %q, got %q", i+1, exp, got)
			return domain.StatusWrongAnswer, domain.VerdictWrongAnswer,
				passed, total, maxRuntimeMS, msg
		}
	}

	return domain.StatusAccepted, domain.VerdictAccepted, passed, total, maxRuntimeMS, ""
}

func expectedOutput(tc domain.TestCase) string {
	if len(tc.ExpectedJSON) > 0 {
		return normalizeOutput(string(tc.ExpectedJSON))
	}
	return normalizeOutput(tc.Expected)
}

func normalizeOutput(value string) string {
	return strings.TrimSpace(value)
}

func outputsEqualWithComparator(got, expected string, comparator domain.ComparatorSpec) bool {
	if got == expected {
		return true
	}

	var gotJSON any
	var expectedJSON any
	if json.Unmarshal([]byte(got), &gotJSON) != nil {
		return false
	}
	if json.Unmarshal([]byte(expected), &expectedJSON) != nil {
		return false
	}
	switch comparator.Kind {
	case "unordered_array":
		return unorderedArraysEqual(gotJSON, expectedJSON)
	case "float_epsilon":
		return floatsEqual(gotJSON, expectedJSON, comparator.Epsilon)
	}
	return reflect.DeepEqual(gotJSON, expectedJSON)
}

func unorderedArraysEqual(got, expected any) bool {
	gotSlice, ok := got.([]any)
	if !ok {
		return false
	}
	expectedSlice, ok := expected.([]any)
	if !ok || len(gotSlice) != len(expectedSlice) {
		return false
	}
	gotValues := canonicalValues(gotSlice)
	expectedValues := canonicalValues(expectedSlice)
	sort.Strings(gotValues)
	sort.Strings(expectedValues)
	return reflect.DeepEqual(gotValues, expectedValues)
}

func canonicalValues(values []any) []string {
	out := make([]string, 0, len(values))
	for _, value := range values {
		encoded, err := json.Marshal(value)
		if err != nil {
			out = append(out, fmt.Sprint(value))
			continue
		}
		out = append(out, string(encoded))
	}
	return out
}

func floatsEqual(got, expected any, epsilon float64) bool {
	if epsilon <= 0 {
		epsilon = 1e-9
	}
	gotFloat, gotOK := got.(float64)
	expectedFloat, expectedOK := expected.(float64)
	return gotOK && expectedOK && math.Abs(gotFloat-expectedFloat) <= epsilon
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
