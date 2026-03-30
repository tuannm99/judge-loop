package main

import (
	"github.com/hibiken/asynq"
	queueadapter "github.com/tuannm99/judge-loop/internal/adapter/queue"
	postgres "github.com/tuannm99/judge-loop/internal/infrastructure/postgres"
	"github.com/tuannm99/judge-loop/internal/infrastructure/queue"
)

// Worker wraps the asynq server and task router.
type Worker struct {
	server *asynq.Server
	mux    *asynq.ServeMux
}

// NewWorker creates a Worker and registers all task handlers.
func NewWorker(cfg Config, db *postgres.DB) *Worker {
	srv := queue.NewServer(cfg.RedisURL, cfg.Concurrency)
	mux := asynq.NewServeMux()

	ev := queueadapter.NewEvaluator(cfg.TimeLimitSecs, db)
	mux.HandleFunc(queue.TypeEvaluateSubmission, ev.ProcessTask)

	return &Worker{server: srv, mux: mux}
}

// Run starts the asynq worker loop (blocks until signal or error).
func (w *Worker) Run() error {
	return w.server.Run(w.mux)
}
