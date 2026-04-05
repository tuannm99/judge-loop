package diserver

import (
	"context"

	"github.com/hibiken/asynq"
	queueadapter "github.com/tuannm99/judge-loop/internal/adapter/queue"
	"github.com/tuannm99/judge-loop/internal/config"
	infraqueue "github.com/tuannm99/judge-loop/internal/infrastructure/queue"
	outport "github.com/tuannm99/judge-loop/internal/port/out"
	"go.uber.org/fx"
)

var queueModule fx.Option = fx.Module("queue",
	fx.Provide(
		provideQueueClient,
		fx.Annotate(queueadapter.NewEvaluationPublisher, fx.As(new(outport.EvaluationPublisher))),
	),
)

func provideQueueClient(lc fx.Lifecycle, cfg config.APIServer) *asynq.Client {
	client := infraqueue.NewClient(cfg.RedisURL)
	lc.Append(fx.Hook{
		OnStop: func(context.Context) error {
			return client.Close()
		},
	})
	return client
}
