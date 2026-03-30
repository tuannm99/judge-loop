package queueadapter

import (
	"context"
	"encoding/json"
	"fmt"
	"log"

	"github.com/google/uuid"
	"github.com/hibiken/asynq"
	sandboxadapter "github.com/tuannm99/judge-loop/internal/adapter/sandbox"
	storageadapter "github.com/tuannm99/judge-loop/internal/adapter/storage"
	application "github.com/tuannm99/judge-loop/internal/application"
	postgres "github.com/tuannm99/judge-loop/internal/infrastructure/postgres"
	q "github.com/tuannm99/judge-loop/internal/infrastructure/queue"
	inport "github.com/tuannm99/judge-loop/internal/port/in"
)

// Evaluator processes submission evaluation jobs from the queue.
type Evaluator struct {
	timeLimitSecs int
	service       inport.EvaluationService
}

// NewEvaluator creates an Evaluator wired to the given database.
func NewEvaluator(timeLimitSecs int, db *postgres.DB) *Evaluator {
	return &Evaluator{
		timeLimitSecs: timeLimitSecs,
		service: application.NewEvaluationService(
			storageadapter.NewSubmissionRepository(db),
			storageadapter.NewTestCaseRepository(db),
			storageadapter.NewSessionRepository(db),
			sandboxadapter.NewRunner(),
		),
	}
}

// ProcessTask implements asynq.Handler for TypeEvaluateSubmission jobs.
func (e *Evaluator) ProcessTask(ctx context.Context, t *asynq.Task) error {
	var payload q.EvaluatePayload
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

	log.Printf("evaluating submission %s", subID)
	return e.service.EvaluateSubmission(ctx, subID, userID, e.timeLimitSecs)
}
