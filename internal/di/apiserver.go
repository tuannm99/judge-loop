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
			provideProblemService,
			provideSubmissionService,
			provideProgressService,
			provideTimerService,
			provideReviewService,
			provideRegistryService,
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

func provideProblemService(db *postgres.DB) inport.ProblemService {
	return application.NewProblemService(postgres.NewProblemStore(db))
}

func provideSubmissionService(db *postgres.DB, queueClient *asynq.Client) inport.SubmissionService {
	return application.NewSubmissionService(
		postgres.NewSubmissionStore(db),
		queueadapter.NewEvaluationPublisher(queueClient),
	)
}

func provideProgressService(db *postgres.DB) inport.ProgressService {
	return application.NewProgressService(postgres.NewSessionStore(db))
}

func provideTimerService(db *postgres.DB) inport.TimerService {
	return application.NewTimerService(postgres.NewSessionStore(db))
}

func provideReviewService(db *postgres.DB) inport.ReviewService {
	return application.NewReviewService(postgres.NewReviewStore(db))
}

func provideRegistryService(db *postgres.DB) inport.RegistryService {
	return application.NewRegistryService(postgres.NewProblemStore(db), postgres.NewRegistryStore(db))
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
