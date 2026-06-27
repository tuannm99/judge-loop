package judge

import (
	"encoding/json"
	"errors"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/tuannm99/judge-loop/internal/domain"
)

func TestEvaluate(t *testing.T) {
	t.Run("accepts when no cases exist", func(t *testing.T) {
		status, verdict, passed, total, runtimeMS, errMsg := Evaluate(nil, nil)
		require.Equal(t, domain.StatusAccepted, status)
		require.Equal(t, domain.VerdictAccepted, verdict)
		require.Zero(t, passed)
		require.Zero(t, total)
		require.Zero(t, runtimeMS)
		require.Empty(t, errMsg)
	})

	t.Run("returns runtime error when runner fails", func(t *testing.T) {
		runErr := errors.New("boom")
		status, verdict, passed, total, _, errMsg := Evaluate(
			[]domain.TestCase{{Input: "1"}},
			func(domain.TestCase) (RunResult, error) {
				return RunResult{}, runErr
			},
		)
		require.Equal(t, domain.StatusRuntimeError, status)
		require.Equal(t, domain.VerdictRuntimeError, verdict)
		require.Zero(t, passed)
		require.Equal(t, 1, total)
		require.Equal(t, runErr.Error(), errMsg)
	})

	t.Run("returns time limit exceeded", func(t *testing.T) {
		status, verdict, passed, total, _, errMsg := Evaluate(
			[]domain.TestCase{{Input: "1"}},
			func(domain.TestCase) (RunResult, error) {
				return RunResult{TimedOut: true}, nil
			},
		)
		require.Equal(t, domain.StatusTimeLimitExceeded, status)
		require.Equal(t, domain.VerdictTimeLimitExceeded, verdict)
		require.Zero(t, passed)
		require.Equal(t, 1, total)
		require.Empty(t, errMsg)
	})

	t.Run("returns compile error", func(t *testing.T) {
		status, verdict, _, _, _, errMsg := Evaluate(
			[]domain.TestCase{{Input: "1"}},
			func(domain.TestCase) (RunResult, error) {
				return RunResult{ExitCode: 1, Stderr: "syntax error near x"}, nil
			},
		)
		require.Equal(t, domain.StatusCompileError, status)
		require.Equal(t, domain.VerdictCompileError, verdict)
		require.Contains(t, errMsg, "syntax error")
	})

	t.Run("returns runtime error for non compile stderr", func(t *testing.T) {
		status, verdict, _, _, _, errMsg := Evaluate(
			[]domain.TestCase{{Input: "1"}},
			func(domain.TestCase) (RunResult, error) {
				return RunResult{ExitCode: 1, Stderr: "panic: nil pointer"}, nil
			},
		)
		require.Equal(t, domain.StatusRuntimeError, status)
		require.Equal(t, domain.VerdictRuntimeError, verdict)
		require.Contains(t, errMsg, "panic")
	})

	t.Run("returns wrong answer and max runtime", func(t *testing.T) {
		cases := []domain.TestCase{{Input: "1", Expected: "1"}, {Input: "2", Expected: "2"}}
		idx := 0
		status, verdict, passed, total, runtimeMS, errMsg := Evaluate(cases, func(domain.TestCase) (RunResult, error) {
			idx++
			if idx == 1 {
				return RunResult{Output: "1", RuntimeMS: 7}, nil
			}
			return RunResult{Output: "3", RuntimeMS: 9}, nil
		})
		require.Equal(t, domain.StatusWrongAnswer, status)
		require.Equal(t, domain.VerdictWrongAnswer, verdict)
		require.Equal(t, 1, passed)
		require.Equal(t, 2, total)
		require.EqualValues(t, 9, runtimeMS)
		require.Contains(t, errMsg, `expected "2", got "3"`)
	})

	t.Run("accepts when all outputs match after trimming", func(t *testing.T) {
		cases := []domain.TestCase{{Input: "1", Expected: " 1 "}, {Input: "2", Expected: "2"}}
		idx := 0
		status, verdict, passed, total, runtimeMS, errMsg := Evaluate(cases, func(domain.TestCase) (RunResult, error) {
			idx++
			return RunResult{Output: "\n" + cases[idx-1].Expected + "\n", RuntimeMS: int64(3 + idx)}, nil
		})
		require.Equal(t, domain.StatusAccepted, status)
		require.Equal(t, domain.VerdictAccepted, verdict)
		require.Equal(t, 2, passed)
		require.Equal(t, 2, total)
		require.EqualValues(t, 5, runtimeMS)
		require.Empty(t, errMsg)
	})

	t.Run("accepts semantically equal json", func(t *testing.T) {
		cases := []domain.TestCase{
			{Input: "array", Expected: "[1,2,3]"},
			{Input: "object", Expected: `{"answer":[1,2],"ok":true}`},
		}
		idx := 0
		status, verdict, passed, total, _, errMsg := Evaluate(cases, func(domain.TestCase) (RunResult, error) {
			idx++
			if idx == 1 {
				return RunResult{Output: "[1, 2, 3]"}, nil
			}
			return RunResult{Output: `{"ok":true,"answer":[1,2]}`}, nil
		})
		require.Equal(t, domain.StatusAccepted, status)
		require.Equal(t, domain.VerdictAccepted, verdict)
		require.Equal(t, 2, passed)
		require.Equal(t, 2, total)
		require.Empty(t, errMsg)
	})

	t.Run("uses expected json when present", func(t *testing.T) {
		status, verdict, passed, total, _, errMsg := Evaluate(
			[]domain.TestCase{{InputJSON: []byte(`{"args":[1]}`), ExpectedJSON: []byte(`[1,2]`)}},
			func(domain.TestCase) (RunResult, error) {
				return RunResult{Output: "[1, 2]"}, nil
			},
		)
		require.Equal(t, domain.StatusAccepted, status)
		require.Equal(t, domain.VerdictAccepted, verdict)
		require.Equal(t, 1, passed)
		require.Equal(t, 1, total)
		require.Empty(t, errMsg)
	})

	t.Run("supports unordered array comparator", func(t *testing.T) {
		status, verdict, passed, total, _, errMsg := EvaluateWithSpec(
			[]domain.TestCase{{ExpectedJSON: []byte(`[1,2]`)}},
			domain.ExecutionSpec{Comparator: domain.ComparatorSpec{Kind: "unordered_array"}},
			func(domain.TestCase) (RunResult, error) {
				return RunResult{Output: "[2,1]"}, nil
			},
		)
		require.Equal(t, domain.StatusAccepted, status)
		require.Equal(t, domain.VerdictAccepted, verdict)
		require.Equal(t, 1, passed)
		require.Equal(t, 1, total)
		require.Empty(t, errMsg)
	})

	t.Run("supports nested unordered and recursive float comparators", func(t *testing.T) {
		status, _, passed, _, _, _ := EvaluateWithSpec(
			[]domain.TestCase{{ExpectedJSON: []byte(`[[1,2],[3]]`)}},
			domain.ExecutionSpec{
				Comparator: domain.ComparatorSpec{Kind: "unordered_nested_array"},
			},
			func(domain.TestCase) (RunResult, error) {
				return RunResult{Output: `[[3],[2,1]]`}, nil
			},
		)
		require.Equal(t, domain.StatusAccepted, status)
		require.Equal(t, 1, passed)

		status, _, passed, _, _, _ = EvaluateWithSpec(
			[]domain.TestCase{{ExpectedJSON: []byte(`[1.0,2.0]`)}},
			domain.ExecutionSpec{
				Comparator: domain.ComparatorSpec{Kind: "float_epsilon", Epsilon: 0.01},
			},
			func(domain.TestCase) (RunResult, error) {
				return RunResult{Output: `[1.001,1.999]`}, nil
			},
		)
		require.Equal(t, domain.StatusAccepted, status)
		require.Equal(t, 1, passed)
	})

	t.Run("validates topological orders", func(t *testing.T) {
		tc := domain.TestCase{
			InputJSON:    []byte(`{"args":[4,[[1,0],[2,0],[3,1],[3,2]]]}`),
			ExpectedJSON: []byte(`[0,1,2,3]`),
		}
		status, _, passed, _, _, _ := EvaluateWithSpec(
			[]domain.TestCase{tc},
			domain.ExecutionSpec{
				Comparator: domain.ComparatorSpec{Kind: "valid_topological_order"},
			},
			func(domain.TestCase) (RunResult, error) {
				return RunResult{Output: `[0,2,1,3]`}, nil
			},
		)
		require.Equal(t, domain.StatusAccepted, status)
		require.Equal(t, 1, passed)
	})

	t.Run("falls back to string comparison for non json", func(t *testing.T) {
		status, verdict, passed, total, _, errMsg := Evaluate(
			[]domain.TestCase{{Input: "plain", Expected: "hello"}},
			func(domain.TestCase) (RunResult, error) {
				return RunResult{Output: "hello world"}, nil
			},
		)
		require.Equal(t, domain.StatusWrongAnswer, status)
		require.Equal(t, domain.VerdictWrongAnswer, verdict)
		require.Zero(t, passed)
		require.Equal(t, 1, total)
		require.Contains(t, errMsg, `expected "hello", got "hello world"`)
	})
}

func TestIsCompileError(t *testing.T) {
	require.True(t, isCompileError("SyntaxError: invalid syntax"))
	require.True(t, isCompileError("undefined: fmt.Println"))
	require.False(t, isCompileError("panic: runtime error"))
}

func TestComparatorEdgeCases(t *testing.T) {
	t.Run("any of", func(t *testing.T) {
		require.True(t, outputsEqualWithComparator(
			`2`,
			`[1,2,3]`,
			domain.ComparatorSpec{Kind: "any_of"},
			domain.TestCase{},
		))
		require.False(t, anyExpectedValueEqual(2.0, map[string]any{}))
		require.False(t, anyExpectedValueEqual(4.0, []any{1.0, 2.0}))
	})

	t.Run("json subset", func(t *testing.T) {
		require.True(t, jsonSubset(
			map[string]any{"a": 1.0, "nested": []any{1.0, 2.0, 3.0}},
			map[string]any{"nested": []any{1.0, 2.0}},
		))
		require.False(t, jsonSubset([]any{1.0}, []any{1.0, 2.0}))
		require.False(t, jsonSubset("value", map[string]any{"a": 1.0}))
		require.False(t, jsonSubset(map[string]any{}, map[string]any{"a": 1.0}))
	})

	t.Run("float epsilon", func(t *testing.T) {
		require.True(t, valuesWithinEpsilon(1.0, 1.0+1e-10, 0))
		require.True(t, valuesWithinEpsilon(
			map[string]any{"a": 1.0},
			map[string]any{"a": 1.001},
			0.01,
		))
		require.False(t, valuesWithinEpsilon([]any{1.0}, []any{1.0, 2.0}, 0.01))
		require.False(t, valuesWithinEpsilon(
			map[string]any{"a": 1.0},
			map[string]any{"a": 1.0, "b": 2.0},
			0.01,
		))
		require.False(t, valuesWithinEpsilon(1.0, "1", 0.01))
	})

	t.Run("unordered arrays reject malformed values", func(t *testing.T) {
		require.False(t, unorderedArraysEqual("not-array", []any{}))
		require.False(t, unorderedArraysEqual([]any{1.0}, "not-array"))
		require.False(t, unorderedNestedArraysEqual([]any{1.0}, []any{1.0}))
		require.Len(t, canonicalValues([]any{make(chan int)}), 1)
	})

	t.Run("topological validation rejects malformed orders", func(t *testing.T) {
		validCase := domain.TestCase{
			InputJSON: json.RawMessage(`{"args":[2,[[1,0]]]}`),
		}
		require.False(t, validTopologicalOrder("not-array", validCase))
		require.False(t, validTopologicalOrder([]any{0.0}, domain.TestCase{Input: `{`}))
		require.False(t, validTopologicalOrder([]any{0.0}, validCase))
		require.False(t, validTopologicalOrder([]any{0.5, 1.0}, validCase))
		require.False(t, validTopologicalOrder([]any{0.0, 2.0}, validCase))
		require.False(t, validTopologicalOrder([]any{0.0, 0.0}, validCase))
		require.False(t, validTopologicalOrder([]any{1.0, 0.0}, validCase))
	})

	t.Run("invalid json is not equal", func(t *testing.T) {
		require.False(t, outputsEqualWithComparator(
			`{`,
			`{}`,
			domain.ComparatorSpec{},
			domain.TestCase{},
		))
		require.False(t, outputsEqualWithComparator(
			`{}`,
			`{`,
			domain.ComparatorSpec{},
			domain.TestCase{},
		))
	})
}
