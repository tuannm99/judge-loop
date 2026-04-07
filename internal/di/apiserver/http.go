package diserver

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/google/uuid"
	apiserver "github.com/tuannm99/judge-loop/internal/adapter/http/apiserver"
	"github.com/tuannm99/judge-loop/internal/config"
	inport "github.com/tuannm99/judge-loop/internal/port/in"
	"go.uber.org/fx"
)

var httpModule fx.Option = fx.Module("http",
	fx.Provide(provideHTTP),
)

func provideHTTP(
	cfg config.APIServer,
	problems inport.ProblemService,
	submissions inport.SubmissionService,
	progress inport.ProgressService,
	timers inport.TimerService,
	reviews inport.ReviewService,
	registry inport.RegistryService,
	missions inport.MissionService,
) (*http.Server, error) {
	userID, err := uuid.Parse(cfg.UserID)
	if err != nil {
		return nil, fmt.Errorf("invalid USER_ID %q: %w", cfg.UserID, err)
	}

	api := apiserver.New(problems, submissions, progress, timers, reviews, registry, missions, userID)
	router := apiserver.NewRouter(api)

	return &http.Server{
		Addr:              ":" + cfg.Port,
		Handler:           router,
		ReadHeaderTimeout: 5 * time.Second,
	}, nil
}

func registerLifecycle(lc fx.Lifecycle, server *http.Server) {
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
