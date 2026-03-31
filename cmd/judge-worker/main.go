// judge-worker consumes submission evaluation jobs from Redis and runs user code
// in Docker sandbox containers. It updates the submission verdict in PostgreSQL.
package main

import (
	"log"

	"github.com/tuannm99/judge-loop/internal/config"
	"github.com/tuannm99/judge-loop/internal/di"
)

func main() {
	cfg, err := config.LoadJudgeWorker()
	if err != nil {
		log.Fatalf("config: %v", err)
	}

	log.Printf("judge-worker starting (concurrency=%d, time_limit=%ds, redis=%s)",
		cfg.Concurrency, cfg.TimeLimitSecs, cfg.RedisURL)

	di.NewJudgeWorker(cfg).Run()
}
