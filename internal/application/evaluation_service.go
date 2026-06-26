package application

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/tuannm99/judge-loop/internal/domain"
	"github.com/tuannm99/judge-loop/internal/domain/judge"
	inport "github.com/tuannm99/judge-loop/internal/port/in"
	outport "github.com/tuannm99/judge-loop/internal/port/out"
)

type EvaluationService struct {
	submissions outport.SubmissionRepository
	problems    outport.ProblemRepository
	testCases   outport.TestCaseRepository
	reviews     outport.ReviewRepository
	sessions    outport.SessionRepository
	runner      outport.CodeRunner
}

var _ inport.EvaluationService = (*EvaluationService)(nil)

func NewEvaluationService(
	submissions outport.SubmissionRepository,
	problems outport.ProblemRepository,
	testCases outport.TestCaseRepository,
	reviews outport.ReviewRepository,
	sessions outport.SessionRepository,
	runner outport.CodeRunner,
) *EvaluationService {
	return &EvaluationService{
		submissions: submissions,
		problems:    problems,
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

	started, err := s.submissions.TryStartEvaluation(ctx, submissionID)
	if err != nil {
		return fmt.Errorf("start evaluation %s: %w", submissionID, err)
	}
	if !started {
		return nil
	}

	cases, err := s.testCases.GetAllByProblem(ctx, sub.ProblemID)
	if err != nil {
		return s.failSubmission(ctx, submissionID, fmt.Sprintf("load test cases: %v", err))
	}
	if len(cases) == 0 {
		return s.failSubmission(ctx, submissionID, "no test cases configured for problem")
	}

	spec, err := s.executionSpec(ctx, sub.ProblemID)
	if err != nil {
		return s.failSubmission(ctx, submissionID, fmt.Sprintf("load execution spec: %v", err))
	}

	status, verdict, passed, total, runtimeMS, errMsg := judge.EvaluateWithSpec(
		cases,
		spec,
		func(tc domain.TestCase) (judge.RunResult, error) {
			runCtx, cancel := context.WithTimeout(ctx, time.Duration(timeLimitSecs)*time.Second)
			defer cancel()
			req, err := buildRunRequest(sub, spec, tc)
			if err != nil {
				return judge.RunResult{}, err
			}
			return s.runner.Run(runCtx, outport.RunRequest{
				Language: req.Language,
				Code:     req.Code,
				Input:    req.Input,
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

	if s.reviews != nil {
		if status == domain.StatusAccepted {
			if err := s.reviews.Upsert(ctx, userID, sub.ProblemID); err != nil {
				return fmt.Errorf("update review schedule: %w", err)
			}
		} else {
			// Regression: failed review → reset schedule to retry tomorrow.
			if err := s.reviews.Reset(ctx, userID, sub.ProblemID); err != nil {
				log.Printf("reset review schedule for %s/%s: %v", userID, sub.ProblemID, err)
			}
		}
	}

	if err := s.sessions.RecordSubmission(ctx, userID, status == domain.StatusAccepted); err != nil {
		log.Printf("record submission stats for user %s: %v", userID, err)
	}
	return nil
}

func (s *EvaluationService) executionSpec(ctx context.Context, problemID uuid.UUID) (domain.ExecutionSpec, error) {
	if s.problems == nil {
		return domain.ExecutionSpec{}, nil
	}
	problem, err := s.problems.GetByID(ctx, problemID)
	if err != nil || problem == nil {
		return domain.ExecutionSpec{}, err
	}
	return problem.ExecutionSpec, nil
}

func buildRunRequest(
	sub *domain.Submission,
	spec domain.ExecutionSpec,
	tc domain.TestCase,
) (outport.RunRequest, error) {
	if spec.Mode == "" || spec.Mode == domain.ExecutionModeStdin {
		return outport.RunRequest{
			Language: string(sub.Language),
			Code:     sub.Code,
			Input:    tc.Input,
		}, nil
	}
	if spec.Mode != domain.ExecutionModeFunction {
		return outport.RunRequest{}, fmt.Errorf("unsupported execution mode: %s", spec.Mode)
	}
	if sub.Language != domain.LanguagePython {
		return outport.RunRequest{}, fmt.Errorf("function mode is not supported for %s yet", sub.Language)
	}
	code, err := renderPythonFunctionHarness(sub.Code, spec, tc)
	if err != nil {
		return outport.RunRequest{}, err
	}
	return outport.RunRequest{Language: string(sub.Language), Code: code}, nil
}

var pythonIdentifierPattern = regexp.MustCompile(`^[A-Za-z_][A-Za-z0-9_]*$`)

func renderPythonFunctionHarness(userCode string, spec domain.ExecutionSpec, tc domain.TestCase) (string, error) {
	entrypoint := strings.TrimSpace(spec.Entrypoint)
	if !pythonIdentifierPattern.MatchString(entrypoint) {
		return "", fmt.Errorf("invalid python entrypoint: %q", spec.Entrypoint)
	}
	className := strings.TrimSpace(spec.ClassName)
	if className == "" {
		className = "Solution"
	}
	if !pythonIdentifierPattern.MatchString(className) {
		return "", fmt.Errorf("invalid python class name: %q", spec.ClassName)
	}
	inputJSON := tc.InputJSON
	if len(inputJSON) == 0 {
		inputJSON = []byte(tc.Input)
	}
	if !json.Valid(inputJSON) {
		return "", errors.New("function mode test case input must be valid json")
	}
	encodedInput := strconv.Quote(string(inputJSON))

	return fmt.Sprintf(`from typing import *

%s

import json as __jl_json

__jl_case = __jl_json.loads(%s)
if isinstance(__jl_case, dict) and "args" in __jl_case:
    __jl_args = __jl_case["args"]
else:
    __jl_args = __jl_case
if not isinstance(__jl_args, list):
    __jl_args = [__jl_args]
__jl_result = %s().%s(*__jl_args)
print(__jl_json.dumps(__jl_result, separators=(",", ":")))
`, userCode, encodedInput, className, entrypoint), nil
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
