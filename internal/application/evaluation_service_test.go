package application

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"github.com/tuannm99/judge-loop/internal/domain"
	"github.com/tuannm99/judge-loop/internal/domain/judge"
	outport "github.com/tuannm99/judge-loop/internal/port/out"
	outmocks "github.com/tuannm99/judge-loop/internal/port/out/mocks"
)

func TestNewEvaluationService(t *testing.T) {
	svc := NewEvaluationService(nil, nil, nil, nil, nil, nil)
	require.NotNil(t, svc)
}

func TestEvaluationServiceEvaluateSubmission(t *testing.T) {
	t.Run("fails when submission missing", func(t *testing.T) {
		submissions := outmocks.NewMockSubmissionRepository(t)
		submissions.EXPECT().GetByID(mock.Anything, mock.Anything).Return(nil, errors.New("missing"))

		svc := NewEvaluationService(submissions, nil, nil, nil, nil, nil)
		err := svc.EvaluateSubmission(context.Background(), uuid.New(), uuid.New(), 1)
		require.Error(t, err)
		require.Contains(t, err.Error(), "get submission")
	})

	t.Run("evaluates and records accepted submission", func(t *testing.T) {
		submissions := outmocks.NewMockSubmissionRepository(t)
		testCases := outmocks.NewMockTestCaseRepository(t)
		reviews := outmocks.NewMockReviewRepository(t)
		sessions := outmocks.NewMockSessionRepository(t)
		runner := outmocks.NewMockCodeRunner(t)

		subID := uuid.New()
		userID := uuid.New()
		problemID := uuid.New()
		sub := &domain.Submission{
			ID:        subID,
			UserID:    userID,
			ProblemID: problemID,
			Language:  domain.LanguagePython,
			Code:      "print(1)",
		}
		cases := []domain.TestCase{{ProblemID: problemID, Input: "1", Expected: "1"}}

		submissions.EXPECT().GetByID(mock.Anything, subID).Return(sub, nil)
		submissions.EXPECT().TryStartEvaluation(mock.Anything, subID).Return(true, nil)
		testCases.EXPECT().GetAllByProblem(mock.Anything, problemID).Return(cases, nil)
		runner.EXPECT().Run(mock.Anything, mock.MatchedBy(func(req any) bool {
			r, ok := req.(outport.RunRequest)
			return ok && r.Input == "1" && r.Language == "python"
		})).Return(judge.RunResult{Output: "1", RuntimeMS: 5}, nil)
		submissions.EXPECT().
			UpdateVerdict(
				mock.Anything, subID, string(domain.StatusAccepted), string(domain.VerdictAccepted),
				1, 1, int64(5), "", mock.Anything,
			).
			Return(nil)
		reviews.EXPECT().Upsert(mock.Anything, userID, problemID).Return(nil)
		sessions.EXPECT().RecordSubmission(mock.Anything, userID, true).Return(nil)

		svc := NewEvaluationService(submissions, nil, testCases, reviews, sessions, runner)
		require.NoError(t, svc.EvaluateSubmission(context.Background(), subID, userID, 1))
	})

	t.Run("evaluates hidden test cases", func(t *testing.T) {
		submissions := outmocks.NewMockSubmissionRepository(t)
		testCases := outmocks.NewMockTestCaseRepository(t)
		reviews := outmocks.NewMockReviewRepository(t)
		sessions := outmocks.NewMockSessionRepository(t)
		runner := outmocks.NewMockCodeRunner(t)

		subID := uuid.New()
		userID := uuid.New()
		problemID := uuid.New()
		sub := &domain.Submission{
			ID:        subID,
			UserID:    userID,
			ProblemID: problemID,
			Language:  domain.LanguagePython,
			Code:      "print(input())",
		}
		cases := []domain.TestCase{
			{ProblemID: problemID, Input: "public", Expected: "public"},
			{ProblemID: problemID, Input: "hidden", Expected: "hidden", IsHidden: true},
		}

		submissions.EXPECT().GetByID(mock.Anything, subID).Return(sub, nil)
		submissions.EXPECT().TryStartEvaluation(mock.Anything, subID).Return(true, nil)
		testCases.EXPECT().GetAllByProblem(mock.Anything, problemID).Return(cases, nil)
		runner.EXPECT().
			Run(mock.Anything, mock.Anything).
			RunAndReturn(func(_ context.Context, req outport.RunRequest) (judge.RunResult, error) {
				return judge.RunResult{Output: req.Input, RuntimeMS: 5}, nil
			})
		submissions.EXPECT().
			UpdateVerdict(
				mock.Anything, subID, string(domain.StatusAccepted), string(domain.VerdictAccepted),
				2, 2, int64(5), "", mock.Anything,
			).
			Return(nil)
		reviews.EXPECT().Upsert(mock.Anything, userID, problemID).Return(nil)
		sessions.EXPECT().RecordSubmission(mock.Anything, userID, true).Return(nil)

		svc := NewEvaluationService(submissions, nil, testCases, reviews, sessions, runner)
		require.NoError(t, svc.EvaluateSubmission(context.Background(), subID, userID, 1))
	})

	t.Run("renders function harness from problem execution spec", func(t *testing.T) {
		submissions := outmocks.NewMockSubmissionRepository(t)
		problems := outmocks.NewMockProblemRepository(t)
		testCases := outmocks.NewMockTestCaseRepository(t)
		reviews := outmocks.NewMockReviewRepository(t)
		sessions := outmocks.NewMockSessionRepository(t)
		runner := outmocks.NewMockCodeRunner(t)

		subID := uuid.New()
		userID := uuid.New()
		problemID := uuid.New()
		sub := &domain.Submission{
			ID:        subID,
			UserID:    userID,
			ProblemID: problemID,
			Language:  domain.LanguagePython,
			Code:      "class Solution:\n    def twoSum(self, nums, target):\n        return [0, 1]",
		}
		problem := &domain.Problem{
			ID: problemID,
			ExecutionSpec: domain.ExecutionSpec{
				Mode:       domain.ExecutionModeFunction,
				Entrypoint: "twoSum",
			},
		}
		cases := []domain.TestCase{
			{
				ProblemID:    problemID,
				InputJSON:    []byte(`{"args":[[2,7],9]}`),
				ExpectedJSON: []byte(`[0,1]`),
			},
		}

		submissions.EXPECT().GetByID(mock.Anything, subID).Return(sub, nil)
		submissions.EXPECT().TryStartEvaluation(mock.Anything, subID).Return(true, nil)
		testCases.EXPECT().GetAllByProblem(mock.Anything, problemID).Return(cases, nil)
		problems.EXPECT().GetByID(mock.Anything, problemID).Return(problem, nil)
		runner.EXPECT().Run(mock.Anything, mock.MatchedBy(func(req outport.RunRequest) bool {
			return req.Language == "python" &&
				req.Input == "" &&
				strings.Contains(req.Code, "Solution().twoSum(*__jl_args)")
		})).Return(judge.RunResult{Output: "[0,1]", RuntimeMS: 5}, nil)
		submissions.EXPECT().
			UpdateVerdict(
				mock.Anything, subID, string(domain.StatusAccepted), string(domain.VerdictAccepted),
				1, 1, int64(5), "", mock.Anything,
			).
			Return(nil)
		reviews.EXPECT().Upsert(mock.Anything, userID, problemID).Return(nil)
		sessions.EXPECT().RecordSubmission(mock.Anything, userID, true).Return(nil)

		svc := NewEvaluationService(submissions, problems, testCases, reviews, sessions, runner)
		require.NoError(t, svc.EvaluateSubmission(context.Background(), subID, userID, 1))
	})

	t.Run("fails closed when test cases cannot be loaded", func(t *testing.T) {
		submissions := outmocks.NewMockSubmissionRepository(t)
		testCases := outmocks.NewMockTestCaseRepository(t)

		subID := uuid.New()
		problemID := uuid.New()
		sub := &domain.Submission{ID: subID, ProblemID: problemID, Language: domain.LanguagePython, Code: "print(1)"}

		submissions.EXPECT().GetByID(mock.Anything, subID).Return(sub, nil)
		submissions.EXPECT().TryStartEvaluation(mock.Anything, subID).Return(true, nil)
		testCases.EXPECT().GetAllByProblem(mock.Anything, problemID).Return(nil, errors.New("no cases"))
		submissions.EXPECT().
			UpdateVerdict(
				mock.Anything, subID, string(domain.StatusRuntimeError), string(domain.VerdictRuntimeError),
				0, 0, int64(0), "load test cases: no cases", mock.Anything,
			).
			Return(nil)

		svc := NewEvaluationService(submissions, nil, testCases, nil, nil, nil)
		err := svc.EvaluateSubmission(context.Background(), subID, uuid.New(), 1)
		require.Error(t, err)
		require.Contains(t, err.Error(), "load test cases")
	})

	t.Run("fails closed when no test cases exist", func(t *testing.T) {
		submissions := outmocks.NewMockSubmissionRepository(t)
		testCases := outmocks.NewMockTestCaseRepository(t)

		subID := uuid.New()
		problemID := uuid.New()
		sub := &domain.Submission{ID: subID, ProblemID: problemID, Language: domain.LanguagePython, Code: "print(1)"}

		submissions.EXPECT().GetByID(mock.Anything, subID).Return(sub, nil)
		submissions.EXPECT().TryStartEvaluation(mock.Anything, subID).Return(true, nil)
		testCases.EXPECT().GetAllByProblem(mock.Anything, problemID).Return([]domain.TestCase{}, nil)
		submissions.EXPECT().
			UpdateVerdict(
				mock.Anything, subID, string(domain.StatusRuntimeError), string(domain.VerdictRuntimeError),
				0, 0, int64(0), "no test cases configured for problem", mock.Anything,
			).
			Return(nil)

		svc := NewEvaluationService(submissions, nil, testCases, nil, nil, nil)
		err := svc.EvaluateSubmission(context.Background(), subID, uuid.New(), 1)
		require.Error(t, err)
		require.Contains(t, err.Error(), "no test cases")
	})

	t.Run("returns review update error", func(t *testing.T) {
		submissions := outmocks.NewMockSubmissionRepository(t)
		testCases := outmocks.NewMockTestCaseRepository(t)
		reviews := outmocks.NewMockReviewRepository(t)
		sessions := outmocks.NewMockSessionRepository(t)
		runner := outmocks.NewMockCodeRunner(t)

		subID := uuid.New()
		userID := uuid.New()
		problemID := uuid.New()
		sub := &domain.Submission{
			ID:        subID,
			UserID:    userID,
			ProblemID: problemID,
			Language:  domain.LanguagePython,
			Code:      "print(1)",
		}
		cases := []domain.TestCase{{ProblemID: problemID, Input: "1", Expected: "1"}}

		submissions.EXPECT().GetByID(mock.Anything, subID).Return(sub, nil)
		submissions.EXPECT().TryStartEvaluation(mock.Anything, subID).Return(true, nil)
		testCases.EXPECT().GetAllByProblem(mock.Anything, problemID).Return(cases, nil)
		runner.EXPECT().Run(mock.Anything, mock.Anything).Return(judge.RunResult{Output: "1", RuntimeMS: 5}, nil)
		submissions.EXPECT().
			UpdateVerdict(
				mock.Anything, subID, string(domain.StatusAccepted), string(domain.VerdictAccepted),
				1, 1, int64(5), "", mock.Anything,
			).
			Return(nil)
		reviews.EXPECT().Upsert(mock.Anything, userID, problemID).Return(errors.New("review failed"))

		svc := NewEvaluationService(submissions, nil, testCases, reviews, sessions, runner)
		err := svc.EvaluateSubmission(context.Background(), subID, userID, 1)
		require.Error(t, err)
		require.Contains(t, err.Error(), "update review schedule")
	})

	t.Run("returns nil when another evaluator already claimed submission", func(t *testing.T) {
		submissions := outmocks.NewMockSubmissionRepository(t)

		subID := uuid.New()
		sub := &domain.Submission{ID: subID, ProblemID: uuid.New(), Language: domain.LanguagePython, Code: "print(1)"}

		submissions.EXPECT().GetByID(mock.Anything, subID).Return(sub, nil)
		submissions.EXPECT().TryStartEvaluation(mock.Anything, subID).Return(false, nil)

		svc := NewEvaluationService(submissions, nil, nil, nil, nil, nil)
		require.NoError(t, svc.EvaluateSubmission(context.Background(), subID, uuid.New(), 1))
	})
}

func TestBuildRunRequestUsesStdinModeByDefault(t *testing.T) {
	sub := &domain.Submission{
		Language: domain.LanguageGo,
		Code:     "package main\nfunc main(){}",
	}
	req, err := buildRunRequest(sub, domain.ExecutionSpec{}, domain.TestCase{Input: "1 2"})
	require.NoError(t, err)
	require.Equal(t, "go", req.Language)
	require.Equal(t, sub.Code, req.Code)
	require.Equal(t, "1 2", req.Input)
}

func TestBuildRunRequestRejectsUnsupportedFunctionLanguage(t *testing.T) {
	sub := &domain.Submission{Language: domain.LanguageGo, Code: "func twoSum() {}"}
	_, err := buildRunRequest(
		sub,
		domain.ExecutionSpec{Mode: domain.ExecutionModeFunction, Entrypoint: "twoSum"},
		domain.TestCase{InputJSON: []byte(`{"args":[[2,7],9]}`)},
	)
	require.Error(t, err)
	require.Contains(t, err.Error(), "function mode is not supported")
}

func TestRenderPythonFunctionHarness(t *testing.T) {
	code, err := renderPythonFunctionHarness(
		"class Solution:\n    def twoSum(self, nums, target):\n        return [0, 1]",
		domain.ExecutionSpec{Mode: domain.ExecutionModeFunction, Entrypoint: "twoSum"},
		domain.TestCase{InputJSON: []byte(`{"args":[[2,7],9]}`)},
	)
	require.NoError(t, err)
	require.Contains(t, code, "Solution().twoSum(*__jl_args)")
	require.Contains(t, code, `__jl_json.loads("{\"args\":[[2,7],9]}")`)
}

func TestRenderPythonFunctionHarnessRejectsInvalidInputJSON(t *testing.T) {
	_, err := renderPythonFunctionHarness(
		"class Solution: pass",
		domain.ExecutionSpec{Mode: domain.ExecutionModeFunction, Entrypoint: "twoSum"},
		domain.TestCase{InputJSON: []byte(`{`)},
	)
	require.Error(t, err)
	require.Contains(t, err.Error(), "valid json")
}
