package dijudge

import (
	"context"

	queueadapter "github.com/tuannm99/judge-loop/internal/adapter/queue"
	"github.com/tuannm99/judge-loop/internal/config"
	postgres "github.com/tuannm99/judge-loop/internal/infrastructure/postgres"
	inport "github.com/tuannm99/judge-loop/internal/port/in"
	outport "github.com/tuannm99/judge-loop/internal/port/out"
	"go.uber.org/fx"
)

var workerModule fx.Option = fx.Module(
	"worker",
	fx.Provide(
		fx.Annotate(postgres.NewEvaluationJobRepositoryImpl, fx.As(new(outport.EvaluationJobQueue))),
		provideEvaluator,
		provideWorker,
	),
)

func provideEvaluator(cfg config.JudgeWorker, service inport.EvaluationService) *queueadapter.Evaluator {
	return queueadapter.NewEvaluator(cfg.TimeLimitSecs, service)
}

func provideWorker(
	cfg config.JudgeWorker,
	queue outport.EvaluationJobQueue,
	evaluator *queueadapter.Evaluator,
) *queueadapter.Worker {
	return queueadapter.NewWorker(queue, evaluator, cfg.WorkerID, cfg.Concurrency)
}

func registerLifecycle(lc fx.Lifecycle, worker *queueadapter.Worker) {
	ctx, cancel := context.WithCancel(context.Background())
	lc.Append(fx.Hook{
		OnStart: func(context.Context) error {
			go worker.Run(ctx)
			return nil
		},
		OnStop: func(context.Context) error {
			cancel()
			return nil
		},
	})
}
