package queueadapter

import (
	"context"
	"fmt"
	"log"

	"github.com/tuannm99/judge-loop/internal/port/in"
	"github.com/tuannm99/judge-loop/internal/port/out"
)

// Evaluator runs a claimed evaluation job through the application service.
type Evaluator struct {
	timeLimitSecs int
	service       in.EvaluationService
}

func NewEvaluator(timeLimitSecs int, service in.EvaluationService) *Evaluator {
	return &Evaluator{
		timeLimitSecs: timeLimitSecs,
		service:       service,
	}
}

func (e *Evaluator) ProcessJob(ctx context.Context, job out.EvaluationJob) error {
	log.Printf("evaluating submission %s", job.SubmissionID)
	if err := e.service.EvaluateSubmission(ctx, job.SubmissionID, job.UserID, e.timeLimitSecs); err != nil {
		return fmt.Errorf("evaluate submission %s: %w", job.SubmissionID, err)
	}
	return nil
}
