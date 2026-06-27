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
		if outputsEqualWithComparator(got, exp, spec.Comparator, tc) {
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

func outputsEqualWithComparator(
	got, expected string,
	comparator domain.ComparatorSpec,
	tc domain.TestCase,
) bool {
	if got == expected && comparator.Kind != "any_of" {
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
	case "unordered_nested_array":
		return unorderedNestedArraysEqual(gotJSON, expectedJSON)
	case "float_epsilon":
		return valuesWithinEpsilon(gotJSON, expectedJSON, comparator.Epsilon)
	case "any_of":
		return anyExpectedValueEqual(gotJSON, expectedJSON)
	case "json_subset":
		return jsonSubset(gotJSON, expectedJSON)
	case "valid_topological_order":
		return validTopologicalOrder(gotJSON, tc)
	}
	return reflect.DeepEqual(gotJSON, expectedJSON)
}

func unorderedNestedArraysEqual(got, expected any) bool {
	gotSlice, gotOK := got.([]any)
	expectedSlice, expectedOK := expected.([]any)
	if !gotOK || !expectedOK || len(gotSlice) != len(expectedSlice) {
		return false
	}
	normalize := func(values []any) ([]string, bool) {
		out := make([]string, len(values))
		for i, value := range values {
			nested, ok := value.([]any)
			if !ok {
				return nil, false
			}
			encoded := canonicalValues(nested)
			sort.Strings(encoded)
			body, err := json.Marshal(encoded)
			if err != nil {
				return nil, false
			}
			out[i] = string(body)
		}
		sort.Strings(out)
		return out, true
	}
	gotValues, gotOK := normalize(gotSlice)
	expectedValues, expectedOK := normalize(expectedSlice)
	return gotOK && expectedOK && reflect.DeepEqual(gotValues, expectedValues)
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

func valuesWithinEpsilon(got, expected any, epsilon float64) bool {
	if epsilon <= 0 {
		epsilon = 1e-9
	}
	switch gotValue := got.(type) {
	case float64:
		expectedValue, ok := expected.(float64)
		return ok && math.Abs(gotValue-expectedValue) <= epsilon
	case []any:
		expectedValue, ok := expected.([]any)
		if !ok || len(gotValue) != len(expectedValue) {
			return false
		}
		for i := range gotValue {
			if !valuesWithinEpsilon(gotValue[i], expectedValue[i], epsilon) {
				return false
			}
		}
		return true
	case map[string]any:
		expectedValue, ok := expected.(map[string]any)
		if !ok || len(gotValue) != len(expectedValue) {
			return false
		}
		for key, value := range gotValue {
			if !valuesWithinEpsilon(value, expectedValue[key], epsilon) {
				return false
			}
		}
		return true
	default:
		return reflect.DeepEqual(got, expected)
	}
}

func anyExpectedValueEqual(got, expected any) bool {
	candidates, ok := expected.([]any)
	if !ok {
		return false
	}
	for _, candidate := range candidates {
		if reflect.DeepEqual(got, candidate) {
			return true
		}
	}
	return false
}

func jsonSubset(got, expected any) bool {
	switch expectedValue := expected.(type) {
	case map[string]any:
		gotValue, ok := got.(map[string]any)
		if !ok {
			return false
		}
		for key, value := range expectedValue {
			if !jsonSubset(gotValue[key], value) {
				return false
			}
		}
		return true
	case []any:
		gotValue, ok := got.([]any)
		if !ok || len(gotValue) < len(expectedValue) {
			return false
		}
		for i := range expectedValue {
			if !jsonSubset(gotValue[i], expectedValue[i]) {
				return false
			}
		}
		return true
	default:
		return reflect.DeepEqual(got, expected)
	}
}

func validTopologicalOrder(got any, tc domain.TestCase) bool {
	order, ok := got.([]any)
	if !ok {
		return false
	}
	raw := tc.InputJSON
	if len(raw) == 0 {
		raw = json.RawMessage(tc.Input)
	}
	var input struct {
		Args []json.RawMessage `json:"args"`
	}
	if json.Unmarshal(raw, &input) != nil || len(input.Args) < 2 {
		return false
	}
	var courseCount int
	var prerequisites [][]int
	if json.Unmarshal(input.Args[0], &courseCount) != nil ||
		json.Unmarshal(input.Args[1], &prerequisites) != nil ||
		len(order) != courseCount {
		return false
	}
	positions := make(map[int]int, len(order))
	for index, rawCourse := range order {
		course, ok := rawCourse.(float64)
		if !ok || course != math.Trunc(course) {
			return false
		}
		value := int(course)
		if value < 0 || value >= courseCount {
			return false
		}
		if _, duplicate := positions[value]; duplicate {
			return false
		}
		positions[value] = index
	}
	for _, pair := range prerequisites {
		if len(pair) != 2 || positions[pair[1]] >= positions[pair[0]] {
			return false
		}
	}
	return true
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
