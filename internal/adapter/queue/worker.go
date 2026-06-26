package queueadapter

import (
	"context"
	"fmt"
	"log"
	"sync"
	"time"

	outport "github.com/tuannm99/judge-loop/internal/port/out"
)

type Worker struct {
	queue        outport.EvaluationJobQueue
	evaluator    *Evaluator
	workerID     string
	concurrency  int
	pollInterval time.Duration
}

func NewWorker(
	queue outport.EvaluationJobQueue,
	evaluator *Evaluator,
	workerID string,
	concurrency int,
) *Worker {
	if concurrency <= 0 {
		concurrency = 1
	}
	if workerID == "" {
		workerID = "judge-worker"
	}
	return &Worker{
		queue:        queue,
		evaluator:    evaluator,
		workerID:     workerID,
		concurrency:  concurrency,
		pollInterval: time.Second,
	}
}

func (w *Worker) Run(ctx context.Context) {
	var wg sync.WaitGroup
	for i := 0; i < w.concurrency; i++ {
		wg.Add(1)
		go func(slot int) {
			defer wg.Done()
			w.runLoop(ctx, slot)
		}(i)
	}
	wg.Wait()
}

func (w *Worker) runLoop(ctx context.Context, slot int) {
	workerID := w.workerID
	if w.concurrency > 1 {
		workerID = fmt.Sprintf("%s-%s-%d", workerID, time.Now().UTC().Format("20060102150405"), slot)
	}

	timer := time.NewTimer(0)
	defer timer.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-timer.C:
		}

		job, err := w.queue.ClaimEvaluationJob(ctx, workerID)
		if err != nil {
			log.Printf("claim evaluation job: %v", err)
			resetTimer(timer, w.pollInterval)
			continue
		}
		if job == nil {
			resetTimer(timer, w.pollInterval)
			continue
		}

		if err := w.evaluator.ProcessJob(ctx, *job); err != nil {
			log.Printf("evaluation job %s failed: %v", job.ID, err)
			if failErr := w.queue.FailEvaluationJob(ctx, job.ID, err.Error()); failErr != nil {
				log.Printf("mark evaluation job %s failed: %v", job.ID, failErr)
			}
			resetTimer(timer, 0)
			continue
		}

		if err := w.queue.CompleteEvaluationJob(ctx, job.ID); err != nil {
			log.Printf("complete evaluation job %s: %v", job.ID, err)
		}
		resetTimer(timer, 0)
	}
}

func resetTimer(timer *time.Timer, d time.Duration) {
	if !timer.Stop() {
		select {
		case <-timer.C:
		default:
		}
	}
	timer.Reset(d)
}
