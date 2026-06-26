// judge-worker consumes submission evaluation jobs from PostgreSQL and runs
// user code in Docker sandbox containers. It updates submission verdicts in
// PostgreSQL.
package main

import (
	"log"

	"github.com/tuannm99/judge-loop/internal/config"
	dijudge "github.com/tuannm99/judge-loop/internal/di/judgeworker"
)

func main() {
	cfg, err := config.LoadJudgeWorker()
	if err != nil {
		log.Fatalf("config: %v", err)
	}

	log.Printf(
		"judge-worker starting (worker_id=%s, concurrency=%d, time_limit=%ds)",
		cfg.WorkerID,
		cfg.Concurrency,
		cfg.TimeLimitSecs,
	)

	dijudge.New(cfg).Run()
}
