package application

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"github.com/tuannm99/judge-loop/internal/domain"
	"github.com/tuannm99/judge-loop/internal/infrastructure/sandbox"
	outport "github.com/tuannm99/judge-loop/internal/port/out"
	outmocks "github.com/tuannm99/judge-loop/internal/port/out/mocks"
)

func TestNewEvaluationService(t *testing.T) {
	svc := NewEvaluationService(nil, nil, nil, nil, nil)
	require.NotNil(t, svc)
}

func TestEvaluationServiceEvaluateSubmission(t *testing.T) {
	t.Run("fails when submission missing", func(t *testing.T) {
		submissions := outmocks.NewMockSubmissionRepository(t)
		submissions.EXPECT().GetByID(mock.Anything, mock.Anything).Return(nil, errors.New("missing"))

		svc := NewEvaluationService(submissions, nil, nil, nil, nil)
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
		submissions.EXPECT().
			UpdateVerdict(mock.Anything, subID, "running", "", 0, 0, int64(0), "", (*time.Time)(nil)).
			Return(nil)
		testCases.EXPECT().GetByProblem(mock.Anything, problemID).Return(cases, nil)
		runner.EXPECT().Run(mock.Anything, mock.MatchedBy(func(req any) bool {
			r, ok := req.(outport.RunRequest)
			return ok && r.Input == "1" && r.Language == "python"
		})).Return(sandbox.RunResult{Output: "1", RuntimeMS: 5}, nil)
		submissions.EXPECT().
			UpdateVerdict(mock.Anything, subID, string(domain.StatusAccepted), string(domain.VerdictAccepted), 1, 1, int64(5), "", mock.Anything).
			Return(nil)
		reviews.EXPECT().Upsert(mock.Anything, userID, problemID).Return(nil)
		sessions.EXPECT().RecordSubmission(mock.Anything, userID, true).Return(nil)

		svc := NewEvaluationService(submissions, testCases, reviews, sessions, runner)
		require.NoError(t, svc.EvaluateSubmission(context.Background(), subID, userID, 1))
	})

	t.Run("fails closed when test cases cannot be loaded", func(t *testing.T) {
		submissions := outmocks.NewMockSubmissionRepository(t)
		testCases := outmocks.NewMockTestCaseRepository(t)

		subID := uuid.New()
		problemID := uuid.New()
		sub := &domain.Submission{ID: subID, ProblemID: problemID, Language: domain.LanguagePython, Code: "print(1)"}

		submissions.EXPECT().GetByID(mock.Anything, subID).Return(sub, nil)
		submissions.EXPECT().
			UpdateVerdict(mock.Anything, subID, "running", "", 0, 0, int64(0), "", (*time.Time)(nil)).
			Return(nil)
		testCases.EXPECT().GetByProblem(mock.Anything, problemID).Return(nil, errors.New("no cases"))
		submissions.EXPECT().
			UpdateVerdict(mock.Anything, subID, string(domain.StatusRuntimeError), string(domain.VerdictRuntimeError), 0, 0, int64(0), "load test cases: no cases", mock.Anything).
			Return(nil)

		svc := NewEvaluationService(submissions, testCases, nil, nil, nil)
		err := svc.EvaluateSubmission(context.Background(), subID, uuid.New(), 1)
		require.Error(t, err)
		require.Contains(t, err.Error(), "load test cases")
	})

	t.Run("fails closed when no visible test cases exist", func(t *testing.T) {
		submissions := outmocks.NewMockSubmissionRepository(t)
		testCases := outmocks.NewMockTestCaseRepository(t)

		subID := uuid.New()
		problemID := uuid.New()
		sub := &domain.Submission{ID: subID, ProblemID: problemID, Language: domain.LanguagePython, Code: "print(1)"}

		submissions.EXPECT().GetByID(mock.Anything, subID).Return(sub, nil)
		submissions.EXPECT().
			UpdateVerdict(mock.Anything, subID, "running", "", 0, 0, int64(0), "", (*time.Time)(nil)).
			Return(nil)
		testCases.EXPECT().GetByProblem(mock.Anything, problemID).Return([]domain.TestCase{}, nil)
		submissions.EXPECT().
			UpdateVerdict(mock.Anything, subID, string(domain.StatusRuntimeError), string(domain.VerdictRuntimeError), 0, 0, int64(0), "no visible test cases configured for problem", mock.Anything).
			Return(nil)

		svc := NewEvaluationService(submissions, testCases, nil, nil, nil)
		err := svc.EvaluateSubmission(context.Background(), subID, uuid.New(), 1)
		require.Error(t, err)
		require.Contains(t, err.Error(), "no visible test cases")
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
		submissions.EXPECT().
			UpdateVerdict(mock.Anything, subID, "running", "", 0, 0, int64(0), "", (*time.Time)(nil)).
			Return(nil)
		testCases.EXPECT().GetByProblem(mock.Anything, problemID).Return(cases, nil)
		runner.EXPECT().Run(mock.Anything, mock.Anything).Return(sandbox.RunResult{Output: "1", RuntimeMS: 5}, nil)
		submissions.EXPECT().
			UpdateVerdict(mock.Anything, subID, string(domain.StatusAccepted), string(domain.VerdictAccepted), 1, 1, int64(5), "", mock.Anything).
			Return(nil)
		reviews.EXPECT().Upsert(mock.Anything, userID, problemID).Return(errors.New("review failed"))

		svc := NewEvaluationService(submissions, testCases, reviews, sessions, runner)
		err := svc.EvaluateSubmission(context.Background(), subID, userID, 1)
		require.Error(t, err)
		require.Contains(t, err.Error(), "update review schedule")
	})
}
