package diagent

import (
	localagent "github.com/tuannm99/judge-loop/internal/adapter/http/localagent"
	"github.com/tuannm99/judge-loop/internal/config"
	"go.uber.org/fx"
	"go.uber.org/fx/fxevent"
)

func New(cfg config.LocalAgent) *fx.App {
	return fx.New(
		fx.WithLogger(func() fxevent.Logger { return fxevent.NopLogger }),
		fx.Supply(cfg),
		fx.Provide(provideHandler),
		httpModule(),
		fx.Invoke(registerLifecycle),
	)
}

func provideHandler(cfg config.LocalAgent) *localagent.Handler {
	client := localagent.NewAPIClient(cfg.ServerURL, cfg.UserID)
	return localagent.NewHandler(client, cfg.UserID, cfg.RegistryPath)
}
