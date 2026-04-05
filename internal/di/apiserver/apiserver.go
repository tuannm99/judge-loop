package diserver

import (
	"context"
	"fmt"
	"time"

	application "github.com/tuannm99/judge-loop/internal/application"
	"github.com/tuannm99/judge-loop/internal/config"
	postgres "github.com/tuannm99/judge-loop/internal/infrastructure/postgres"
	inport "github.com/tuannm99/judge-loop/internal/port/in"
	outport "github.com/tuannm99/judge-loop/internal/port/out"
	"go.uber.org/fx"
	"go.uber.org/fx/fxevent"
)

func New(cfg config.APIServer) *fx.App {
	return fx.New(
		fx.WithLogger(func() fxevent.Logger { return fxevent.NopLogger }),
		fx.Supply(cfg),
		fx.Provide(provideDB),
		fx.Module("problem",
			fx.Provide(
				fx.Annotate(postgres.NewProblemRepositoryImpl, fx.As(new(outport.ProblemRepository))),
				fx.Annotate(application.NewProblemService, fx.As(new(inport.ProblemService))),
			),
		),
		fx.Module("submission",
			fx.Provide(
				fx.Annotate(postgres.NewSubmissionRepositoryImpl, fx.As(new(outport.SubmissionRepository))),
				fx.Annotate(application.NewSubmissionService, fx.As(new(inport.SubmissionService))),
			),
		),
		fx.Module("session",
			fx.Provide(
				fx.Annotate(postgres.NewSessionRepositoryImpl, fx.As(new(outport.SessionRepository))),
				fx.Annotate(application.NewProgressService, fx.As(new(inport.ProgressService))),
				fx.Annotate(application.NewTimerService, fx.As(new(inport.TimerService))),
			),
		),
		fx.Module("review",
			fx.Provide(
				fx.Annotate(postgres.NewReviewRepositoryImpl, fx.As(new(outport.ReviewRepository))),
				fx.Annotate(application.NewReviewService, fx.As(new(inport.ReviewService))),
			),
		),
		fx.Module("registry",
			fx.Provide(
				fx.Annotate(postgres.NewRegistryRepositoryImpl, fx.As(new(outport.RegistryRepository))),
				fx.Annotate(application.NewRegistryService, fx.As(new(inport.RegistryService))),
			),
		),
		queueModule,
		httpModule,
		fx.Invoke(registerLifecycle),
	)
}

func provideDB(lc fx.Lifecycle, cfg config.APIServer) (*postgres.DB, error) {
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
