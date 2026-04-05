package dijudge

import (
	"context"
	"log"

	"github.com/hibiken/asynq"
	queueadapter "github.com/tuannm99/judge-loop/internal/adapter/queue"
	"github.com/tuannm99/judge-loop/internal/config"
	infraqueue "github.com/tuannm99/judge-loop/internal/infrastructure/queue"
	inport "github.com/tuannm99/judge-loop/internal/port/in"
	"go.uber.org/fx"
)

var workerModule fx.Option = fx.Module("worker",
	fx.Provide(
		provideEvaluator,
		provideServer,
		provideMux,
	),
)

func provideEvaluator(cfg config.JudgeWorker, service inport.EvaluationService) *queueadapter.Evaluator {
	return queueadapter.NewEvaluator(cfg.TimeLimitSecs, service)
}

func provideServer(cfg config.JudgeWorker) *asynq.Server {
	return infraqueue.NewServer(cfg.RedisURL, cfg.Concurrency)
}

func provideMux(evaluator *queueadapter.Evaluator) *asynq.ServeMux {
	mux := asynq.NewServeMux()
	mux.HandleFunc(infraqueue.TypeEvaluateSubmission, evaluator.ProcessTask)
	return mux
}

func registerLifecycle(lc fx.Lifecycle, server *asynq.Server, mux *asynq.ServeMux) {
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
