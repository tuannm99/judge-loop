package diserver

import (
	postgres "github.com/tuannm99/judge-loop/internal/infrastructure/postgres"
	outport "github.com/tuannm99/judge-loop/internal/port/out"
	"go.uber.org/fx"
)

var queueModule fx.Option = fx.Module(
	"queue",
	fx.Provide(
		fx.Annotate(postgres.NewEvaluationJobRepositoryImpl, fx.As(new(outport.EvaluationPublisher))),
	),
)
