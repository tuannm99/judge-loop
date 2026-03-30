package sandboxadapter

import (
	"context"

	"github.com/tuannm99/judge-loop/internal/infrastructure/sandbox"
	outport "github.com/tuannm99/judge-loop/internal/port/out"
)

type Runner struct{}

func NewRunner() *Runner { return &Runner{} }

func (r *Runner) Run(ctx context.Context, req outport.RunRequest) (sandbox.RunResult, error) {
	return sandbox.Run(ctx, sandbox.RunRequest{
		Language: req.Language,
		Code:     req.Code,
		Input:    req.Input,
	})
}
