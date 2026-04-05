// judge-worker consumes submission evaluation jobs from Redis and runs user code
// in Docker sandbox containers. It updates the submission verdict in PostgreSQL.
package main

import (
	"flag"
	"log"
	"net/http"
	"strings"

	"github.com/hibiken/asynq"
	"github.com/hibiken/asynqmon"
	"github.com/tuannm99/judge-loop/internal/config"
	dijudge "github.com/tuannm99/judge-loop/internal/di/judgeworker"
)

func main() {
	enableUI := flag.Bool("ui", false, "enable embedded Asynq monitor UI")
	uiPort := flag.String("ui-port", "8081", "HTTP port for embedded Asynq monitor UI")
	uiPath := flag.String("ui-path", "/monitoring", "root path for embedded Asynq monitor UI")
	uiReadOnly := flag.Bool("ui-readonly", false, "run embedded Asynq monitor UI in read-only mode")
	flag.Parse()

	cfg, err := config.LoadJudgeWorker()
	if err != nil {
		log.Fatalf("config: %v", err)
	}

	log.Printf("judge-worker starting (concurrency=%d, time_limit=%ds, redis=%s)",
		cfg.Concurrency, cfg.TimeLimitSecs, cfg.RedisURL)

	if *enableUI {
		rootPath := *uiPath
		if !strings.HasPrefix(rootPath, "/") {
			rootPath = "/" + rootPath
		}

		h := asynqmon.New(asynqmon.Options{
			RootPath:     rootPath,
			RedisConnOpt: asynq.RedisClientOpt{Addr: cfg.RedisURL},
			ReadOnly:     *uiReadOnly,
		})
		defer func() {
			if err := h.Close(); err != nil {
				log.Printf("asynqmon close: %v", err)
			}
		}()

		go func() {
			mux := http.NewServeMux()
			mux.Handle(rootPath, h)
			mux.Handle(rootPath+"/", h)
			log.Printf("judge-worker ui listening on :%s%s", *uiPort, rootPath)
			if err := http.ListenAndServe(":"+*uiPort, mux); err != nil {
				log.Printf("judge-worker ui: %v", err)
			}
		}()
	}

	dijudge.New(cfg).Run()
}
