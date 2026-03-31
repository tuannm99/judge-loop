package application

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/tuannm99/judge-loop/internal/domain"
	"github.com/tuannm99/judge-loop/internal/domain/judge"
	"github.com/tuannm99/judge-loop/internal/infrastructure/sandbox"
	inport "github.com/tuannm99/judge-loop/internal/port/in"
	outport "github.com/tuannm99/judge-loop/internal/port/out"
)

type EvaluationService struct {
	submissions outport.SubmissionRepository
	testCases   outport.TestCaseRepository
	reviews     outport.ReviewRepository
	sessions    outport.SessionRepository
	runner      outport.CodeRunner
}

var _ inport.EvaluationService = (*EvaluationService)(nil)

func NewEvaluationService(
	submissions outport.SubmissionRepository,
	testCases outport.TestCaseRepository,
	reviews outport.ReviewRepository,
	sessions outport.SessionRepository,
	runner outport.CodeRunner,
) *EvaluationService {
	return &EvaluationService{
		submissions: submissions,
		testCases:   testCases,
		reviews:     reviews,
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
		return s.failSubmission(ctx, submissionID, fmt.Sprintf("load test cases: %v", err))
	}
	if len(cases) == 0 {
		return s.failSubmission(ctx, submissionID, "no visible test cases configured for problem")
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

	if status == domain.StatusAccepted && s.reviews != nil {
		if err := s.reviews.Upsert(ctx, userID, sub.ProblemID); err != nil {
			return fmt.Errorf("update review schedule: %w", err)
		}
	}

	_ = s.sessions.RecordSubmission(ctx, userID, status == domain.StatusAccepted)
	return nil
}

func (s *EvaluationService) failSubmission(ctx context.Context, submissionID uuid.UUID, errMsg string) error {
	now := time.Now()
	if err := s.submissions.UpdateVerdict(
		ctx,
		submissionID,
		string(domain.StatusRuntimeError),
		string(domain.VerdictRuntimeError),
		0,
		0,
		0,
		errMsg,
		&now,
	); err != nil {
		return fmt.Errorf("update verdict: %w", err)
	}
	return errors.New(errMsg)
}
