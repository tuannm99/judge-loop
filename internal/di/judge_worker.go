package di

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/hibiken/asynq"
	queueadapter "github.com/tuannm99/judge-loop/internal/adapter/queue"
	sandboxadapter "github.com/tuannm99/judge-loop/internal/adapter/sandbox"
	application "github.com/tuannm99/judge-loop/internal/application"
	"github.com/tuannm99/judge-loop/internal/config"
	postgres "github.com/tuannm99/judge-loop/internal/infrastructure/postgres"
	infraqueue "github.com/tuannm99/judge-loop/internal/infrastructure/queue"
	inport "github.com/tuannm99/judge-loop/internal/port/in"
	"go.uber.org/fx"
	"go.uber.org/fx/fxevent"
)

func NewJudgeWorker(cfg config.JudgeWorker) *fx.App {
	return fx.New(
		fx.WithLogger(func() fxevent.Logger { return fxevent.NopLogger }),
		fx.Supply(cfg),
		fx.Provide(
			provideJudgeWorkerDB,
			provideEvaluationService,
			provideEvaluator,
			provideWorkerServer,
			provideWorkerMux,
		),
		fx.Invoke(registerWorkerLifecycle),
	)
}

func provideJudgeWorkerDB(lc fx.Lifecycle, cfg config.JudgeWorker) (*postgres.DB, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	db, err := postgres.Connect(ctx, cfg.DatabaseURL)
	if err != nil {
		return nil, fmt.Errorf("db connect: %w", err)
	}

	lc.Append(fx.Hook{
		OnStop: func(context.Context) error {
			db.Close()
			return nil
		},
	})
	return db, nil
}

func provideEvaluationService(db *postgres.DB) inport.EvaluationService {
	return application.NewEvaluationService(
		postgres.NewSubmissionStore(db),
		postgres.NewTestCaseStore(db),
		postgres.NewReviewStore(db),
		postgres.NewSessionStore(db),
		sandboxadapter.NewRunner(),
	)
}

func provideEvaluator(cfg config.JudgeWorker, service inport.EvaluationService) *queueadapter.Evaluator {
	return queueadapter.NewEvaluator(cfg.TimeLimitSecs, service)
}

func provideWorkerServer(cfg config.JudgeWorker) *asynq.Server {
	return infraqueue.NewServer(cfg.RedisURL, cfg.Concurrency)
}

func provideWorkerMux(evaluator *queueadapter.Evaluator) *asynq.ServeMux {
	mux := asynq.NewServeMux()
	mux.HandleFunc(infraqueue.TypeEvaluateSubmission, evaluator.ProcessTask)
	return mux
}

func registerWorkerLifecycle(lc fx.Lifecycle, server *asynq.Server, mux *asynq.ServeMux) {
	lc.Append(fx.Hook{
		OnStart: func(context.Context) error {
			go func() {
				if err := server.Run(mux); err != nil {
					log.Printf("judge-worker: %v", err)
				}
			}()
			return nil
		},
		OnStop: func(context.Context) error {
			server.Shutdown()
			return nil
		},
	})
}
