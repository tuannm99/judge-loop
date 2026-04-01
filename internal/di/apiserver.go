package di

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/hibiken/asynq"
	apiserver "github.com/tuannm99/judge-loop/internal/adapter/http/apiserver"
	queueadapter "github.com/tuannm99/judge-loop/internal/adapter/queue"
	application "github.com/tuannm99/judge-loop/internal/application"
	"github.com/tuannm99/judge-loop/internal/config"
	postgres "github.com/tuannm99/judge-loop/internal/infrastructure/postgres"
	infraqueue "github.com/tuannm99/judge-loop/internal/infrastructure/queue"
	inport "github.com/tuannm99/judge-loop/internal/port/in"
	outport "github.com/tuannm99/judge-loop/internal/port/out"
	"go.uber.org/fx"
	"go.uber.org/fx/fxevent"
)

func NewAPIServer(cfg config.APIServer) *fx.App {
	return fx.New(
		fx.WithLogger(func() fxevent.Logger { return fxevent.NopLogger }),
		fx.Supply(cfg),
		fx.Provide(
			provideAPIServerDB,
			provideAPIQueueClient,
			fx.Annotate(postgres.NewProblemStore, fx.As(new(outport.ProblemRepository))),
			fx.Annotate(postgres.NewSubmissionStore, fx.As(new(outport.SubmissionRepository))),
			fx.Annotate(postgres.NewSessionStore, fx.As(new(outport.SessionRepository))),
			fx.Annotate(postgres.NewReviewStore, fx.As(new(outport.ReviewRepository))),
			fx.Annotate(postgres.NewRegistryStore, fx.As(new(outport.RegistryRepository))),
			fx.Annotate(queueadapter.NewEvaluationPublisher, fx.As(new(outport.EvaluationPublisher))),
			fx.Annotate(application.NewProblemService, fx.As(new(inport.ProblemService))),
			fx.Annotate(application.NewSubmissionService, fx.As(new(inport.SubmissionService))),
			fx.Annotate(application.NewProgressService, fx.As(new(inport.ProgressService))),
			fx.Annotate(application.NewTimerService, fx.As(new(inport.TimerService))),
			fx.Annotate(application.NewReviewService, fx.As(new(inport.ReviewService))),
			fx.Annotate(application.NewRegistryService, fx.As(new(inport.RegistryService))),
			provideAPI,
		),
		fx.Invoke(registerAPIServerLifecycle),
	)
}

func provideAPIServerDB(lc fx.Lifecycle, cfg config.APIServer) (*postgres.DB, error) {
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

func provideAPIQueueClient(lc fx.Lifecycle, cfg config.APIServer) *asynq.Client {
	client := infraqueue.NewClient(cfg.RedisURL)
	lc.Append(fx.Hook{
		OnStop: func(context.Context) error {
			return client.Close()
		},
	})
	return client
}

func provideAPI(
	cfg config.APIServer,
	problems inport.ProblemService,
	submissions inport.SubmissionService,
	progress inport.ProgressService,
	timers inport.TimerService,
	reviews inport.ReviewService,
	registry inport.RegistryService,
) (*http.Server, error) {
	userID, err := uuid.Parse(cfg.UserID)
	if err != nil {
		return nil, fmt.Errorf("invalid USER_ID %q: %w", cfg.UserID, err)
	}

	api := apiserver.New(problems, submissions, progress, timers, reviews, registry, userID)
	router := apiserver.NewRouter(api)

	return &http.Server{
		Addr:              ":" + cfg.Port,
		Handler:           router,
		ReadHeaderTimeout: 5 * time.Second,
	}, nil
}

func registerAPIServerLifecycle(lc fx.Lifecycle, server *http.Server) {
	lc.Append(fx.Hook{
		OnStart: func(context.Context) error {
			go func() {
				if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
					log.Printf("api-server: %v", err)
				}
			}()
			return nil
		},
		OnStop: func(ctx context.Context) error {
			return server.Shutdown(ctx)
		},
	})
}
