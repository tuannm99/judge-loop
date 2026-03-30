// judge-worker consumes submission evaluation jobs from Redis and runs user code
// in Docker sandbox containers. It updates the submission verdict in PostgreSQL.
package main

import (
	"context"
	"log"
	"time"

	"github.com/tuannm99/judge-loop/internal/storage"
)

func main() {
	cfg := LoadConfig()

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	db, err := storage.Connect(ctx, cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("db connect: %v", err)
	}
	defer db.Close()

	w := NewWorker(cfg, db)
	log.Printf("judge-worker starting (concurrency=%d, time_limit=%ds, redis=%s)",
		cfg.Concurrency, cfg.TimeLimitSecs, cfg.RedisURL)

	if err := w.Run(); err != nil {
		log.Fatalf("worker: %v", err)
	}
}
