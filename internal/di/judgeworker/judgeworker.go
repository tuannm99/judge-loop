package dijudge

import (
	"context"
	"fmt"
	"time"

	sandboxadapter "github.com/tuannm99/judge-loop/internal/adapter/sandbox"
	application "github.com/tuannm99/judge-loop/internal/application"
	"github.com/tuannm99/judge-loop/internal/config"
	postgres "github.com/tuannm99/judge-loop/internal/infrastructure/postgres"
	inport "github.com/tuannm99/judge-loop/internal/port/in"
	outport "github.com/tuannm99/judge-loop/internal/port/out"
	"go.uber.org/fx"
	"go.uber.org/fx/fxevent"
)

func New(cfg config.JudgeWorker) *fx.App {
	return fx.New(
		fx.WithLogger(func() fxevent.Logger { return fxevent.NopLogger }),
		fx.Supply(cfg),
		fx.Provide(provideDB),
		fx.Module("evaluation",
			fx.Provide(
				fx.Annotate(postgres.NewSubmissionRepositoryImpl, fx.As(new(outport.SubmissionRepository))),
				fx.Annotate(postgres.NewTestCaseRepositoryImpl, fx.As(new(outport.TestCaseRepository))),
				fx.Annotate(postgres.NewReviewRepositoryImpl, fx.As(new(outport.ReviewRepository))),
				fx.Annotate(postgres.NewSessionRepositoryImpl, fx.As(new(outport.SessionRepository))),
				fx.Annotate(sandboxadapter.NewRunner, fx.As(new(outport.CodeRunner))),
				fx.Annotate(application.NewEvaluationService, fx.As(new(inport.EvaluationService))),
			),
		),
		workerModule,
		fx.Invoke(registerLifecycle),
	)
}

func provideDB(lc fx.Lifecycle, cfg config.JudgeWorker) (*postgres.DB, error) {
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
