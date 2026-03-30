// api-server is the main HTTP API server for judge-loop.
// It handles problems, submissions, progress, timers, and daily reviews.
package main

import (
	"context"
	"log"
	"time"

	"github.com/tuannm99/judge-loop/internal/queue"
	"github.com/tuannm99/judge-loop/internal/storage"
)

func main() {
	cfg, err := LoadConfig()
	if err != nil {
		log.Fatalf("config: %v", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	db, err := storage.Connect(ctx, cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("db connect: %v", err)
	}
	defer db.Close()

	queueClient := queue.NewClient(cfg.RedisURL)
	defer queueClient.Close()

	srv := NewServer(cfg, db, queueClient)
	log.Printf("api-server listening on :%s (redis: %s)", cfg.Port, cfg.RedisURL)
	if err := srv.Run(); err != nil {
		log.Fatalf("server: %v", err)
	}
}
