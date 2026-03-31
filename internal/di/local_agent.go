package di

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net/http"
	"time"

	localagent "github.com/tuannm99/judge-loop/internal/adapter/http/localagent"
	"github.com/tuannm99/judge-loop/internal/config"
	"go.uber.org/fx"
	"go.uber.org/fx/fxevent"
)

func NewLocalAgent(cfg config.LocalAgent) *fx.App {
	return fx.New(
		fx.WithLogger(func() fxevent.Logger { return fxevent.NopLogger }),
		fx.Supply(cfg),
		fx.Provide(
			provideLocalAgentHandler,
			provideLocalAgentHTTP,
		),
		fx.Invoke(registerLocalAgentLifecycle),
	)
}

func provideLocalAgentHandler(cfg config.LocalAgent) *localagent.Handler {
	client := localagent.NewAPIClient(cfg.ServerURL, cfg.UserID)
	return localagent.NewHandler(client, cfg.UserID, cfg.RegistryPath)
}

func provideLocalAgentHTTP(cfg config.LocalAgent, handler *localagent.Handler) *http.Server {
	return &http.Server{
		Addr:              fmt.Sprintf("127.0.0.1:%d", cfg.Port),
		Handler:           localagent.NewRouter(handler),
		ReadHeaderTimeout: 5 * time.Second,
	}
}

func registerLocalAgentLifecycle(lc fx.Lifecycle, server *http.Server) {
	lc.Append(fx.Hook{
		OnStart: func(context.Context) error {
			go func() {
				if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
					log.Printf("local-agent: %v", err)
				}
			}()
			return nil
		},
		OnStop: func(ctx context.Context) error {
			return server.Shutdown(ctx)
		},
	})
}
