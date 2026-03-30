package application

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/tuannm99/judge-loop/internal/domain"
	"github.com/tuannm99/judge-loop/internal/domain/judge"
	"github.com/tuannm99/judge-loop/internal/infrastructure/sandbox"
	outport "github.com/tuannm99/judge-loop/internal/port/out"
)

type EvaluationService struct {
	submissions outport.SubmissionRepository
	testCases   outport.TestCaseRepository
	sessions    outport.SessionRepository
	runner      outport.CodeRunner
}

func NewEvaluationService(
	submissions outport.SubmissionRepository,
	testCases outport.TestCaseRepository,
	sessions outport.SessionRepository,
	runner outport.CodeRunner,
) *EvaluationService {
	return &EvaluationService{
		submissions: submissions,
		testCases:   testCases,
		sessions:    sessions,
		runner:      runner,
	}
}

func (s *EvaluationService) EvaluateSubmission(
	ctx context.Context,
	submissionID, userID uuid.UUID,
	timeLimitSecs int,
) error {
	sub, err := s.submissions.GetByID(ctx, submissionID)
	if err != nil || sub == nil {
		return fmt.Errorf("get submission %s: %w", submissionID, err)
	}

	_ = s.submissions.UpdateVerdict(ctx, submissionID, "running", "", 0, 0, 0, "", nil)

	cases, err := s.testCases.GetByProblem(ctx, sub.ProblemID)
	if err != nil {
		cases = nil
	}

	status, verdict, passed, total, runtimeMS, errMsg := judge.Evaluate(
		cases,
		func(input string) (sandbox.RunResult, error) {
			runCtx, cancel := context.WithTimeout(ctx, time.Duration(timeLimitSecs)*time.Second)
			defer cancel()
			return s.runner.Run(runCtx, outport.RunRequest{
				Language: string(sub.Language),
				Code:     sub.Code,
				Input:    input,
			})
		},
	)

	now := time.Now()
	if err := s.submissions.UpdateVerdict(
		ctx, submissionID,
		string(status), string(verdict),
		passed, total, runtimeMS, errMsg, &now,
	); err != nil {
		return fmt.Errorf("update verdict: %w", err)
	}

	_ = s.sessions.RecordSubmission(ctx, userID, status == domain.StatusAccepted)
	return nil
}
