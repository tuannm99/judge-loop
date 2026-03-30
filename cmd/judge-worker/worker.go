package main

import (
	"github.com/hibiken/asynq"
	"github.com/tuannm99/judge-loop/internal/queue"
	"github.com/tuannm99/judge-loop/internal/storage"
)

// Worker wraps the asynq server and task router.
type Worker struct {
	server *asynq.Server
	mux    *asynq.ServeMux
}

// NewWorker creates a Worker and registers all task handlers.
func NewWorker(cfg Config, db *storage.DB) *Worker {
	srv := queue.NewServer(cfg.RedisURL, cfg.Concurrency)
	mux := asynq.NewServeMux()

	ev := NewEvaluator(cfg, db)
	mux.HandleFunc(queue.TypeEvaluateSubmission, ev.ProcessTask)

	return &Worker{server: srv, mux: mux}
}

// Run starts the asynq worker loop (blocks until signal or error).
func (w *Worker) Run() error {
	return w.server.Run(w.mux)
}
