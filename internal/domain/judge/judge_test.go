package judge

import (
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
			func(string) (RunResult, error) {
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
			func(string) (RunResult, error) {
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
			func(string) (RunResult, error) {
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
			func(string) (RunResult, error) {
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
		status, verdict, passed, total, runtimeMS, errMsg := Evaluate(cases, func(string) (RunResult, error) {
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
		status, verdict, passed, total, runtimeMS, errMsg := Evaluate(cases, func(string) (RunResult, error) {
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
}

func TestIsCompileError(t *testing.T) {
	require.True(t, isCompileError("SyntaxError: invalid syntax"))
	require.True(t, isCompileError("undefined: fmt.Println"))
	require.False(t, isCompileError("panic: runtime error"))
}
