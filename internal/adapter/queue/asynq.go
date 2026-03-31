package queueadapter

import (
	"log"

	"github.com/hibiken/asynq"
	q "github.com/tuannm99/judge-loop/internal/infrastructure/queue"
	outport "github.com/tuannm99/judge-loop/internal/port/out"
)

type EvaluationPublisher struct {
	client *asynq.Client
}

var _ outport.EvaluationPublisher = (*EvaluationPublisher)(nil)

func NewEvaluationPublisher(client *asynq.Client) *EvaluationPublisher {
	return &EvaluationPublisher{client: client}
}

func (p *EvaluationPublisher) PublishEvaluation(job outport.EvaluateSubmissionJob) error {
	task, err := q.NewEvaluateTask(q.EvaluatePayload{
		SubmissionID: job.SubmissionID,
		UserID:       job.UserID,
	})
	if err != nil {
		return err
	}
	_, err = p.client.Enqueue(task)
	if err != nil {
		log.Printf("enqueue evaluate task for %s: %v", job.SubmissionID, err)
	}
	return err
}
