package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/google/uuid"
	"github.com/hibiken/asynq"
	"github.com/tuannm99/judge-loop/internal/domain"
	"github.com/tuannm99/judge-loop/internal/judge"
	"github.com/tuannm99/judge-loop/internal/queue"
	"github.com/tuannm99/judge-loop/internal/sandbox"
	"github.com/tuannm99/judge-loop/internal/storage"
)

// Evaluator processes submission evaluation jobs from the queue.
type Evaluator struct {
	cfg         Config
	submissions *storage.SubmissionStore
	testCases   *storage.TestCaseStore
	sessions    *storage.SessionStore
}

// NewEvaluator creates an Evaluator wired to the given database.
func NewEvaluator(cfg Config, db *storage.DB) *Evaluator {
	return &Evaluator{
		cfg:         cfg,
		submissions: storage.NewSubmissionStore(db),
		testCases:   storage.NewTestCaseStore(db),
		sessions:    storage.NewSessionStore(db),
	}
}

// ProcessTask implements asynq.Handler for TypeEvaluateSubmission jobs.
func (e *Evaluator) ProcessTask(ctx context.Context, t *asynq.Task) error {
	var payload queue.EvaluatePayload
	if err := json.Unmarshal(t.Payload(), &payload); err != nil {
		return fmt.Errorf("unmarshal payload: %w", err)
	}

	subID, err := uuid.Parse(payload.SubmissionID)
	if err != nil {
		return fmt.Errorf("parse submission_id: %w", err)
	}
	userID, err := uuid.Parse(payload.UserID)
	if err != nil {
		return fmt.Errorf("parse user_id: %w", err)
	}

	sub, err := e.submissions.GetByID(ctx, subID)
	if err != nil || sub == nil {
		return fmt.Errorf("get submission %s: %w", subID, err)
	}

	log.Printf("evaluating submission %s (lang=%s, problem=%s)", subID, sub.Language, sub.ProblemID)

	// Mark as running so the plugin can show "running…" state.
	_ = e.submissions.UpdateVerdict(ctx, subID, "running", "", 0, 0, 0, "", nil)

	cases, err := e.testCases.GetByProblem(ctx, sub.ProblemID)
	if err != nil {
		log.Printf("get test cases for %s: %v (proceeding with 0 cases)", sub.ProblemID, err)
		cases = nil
	}

	timeLimitSecs := e.cfg.TimeLimitSecs
	lang := string(sub.Language)
	code := sub.Code

	status, verdict, passed, total, runtimeMS, errMsg := judge.Evaluate(
		cases,
		func(input string) (sandbox.RunResult, error) {
			runCtx, cancel := context.WithTimeout(ctx, time.Duration(timeLimitSecs)*time.Second)
			defer cancel()
			return sandbox.Run(runCtx, sandbox.RunRequest{
				Language: lang,
				Code:     code,
				Input:    input,
			})
		},
	)

	now := time.Now()
	if err := e.submissions.UpdateVerdict(
		ctx, subID,
		string(status), string(verdict),
		passed, total, runtimeMS, errMsg, &now,
	); err != nil {
		return fmt.Errorf("update verdict: %w", err)
	}

	_ = e.sessions.RecordSubmission(ctx, userID, status == domain.StatusAccepted)

	log.Printf("submission %s → %s (%d/%d cases, %dms)", subID, verdict, passed, total, runtimeMS)
	return nil
}
