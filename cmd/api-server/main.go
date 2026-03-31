// api-server is the main HTTP API server for judge-loop.
// It handles problems, submissions, progress, timers, and daily reviews.
package main

import (
	"log"

	"github.com/tuannm99/judge-loop/internal/config"
	"github.com/tuannm99/judge-loop/internal/di"
)

func main() {
	cfg, err := config.LoadAPIServer()
	if err != nil {
		log.Fatalf("config: %v", err)
	}

	log.Printf("api-server listening on :%s (redis: %s)", cfg.Port, cfg.RedisURL)
	di.NewAPIServer(cfg).Run()
}
