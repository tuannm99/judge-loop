package sandboxadapter

import (
	"context"

	"github.com/tuannm99/judge-loop/internal/domain/judge"
	"github.com/tuannm99/judge-loop/internal/infrastructure/sandbox"
	outport "github.com/tuannm99/judge-loop/internal/port/out"
)

type Runner struct{}

var _ outport.CodeRunner = (*Runner)(nil)

func NewRunner() *Runner { return &Runner{} }

func (r *Runner) Run(ctx context.Context, req outport.RunRequest) (judge.RunResult, error) {
	result, err := sandbox.Run(ctx, sandbox.RunRequest{
		Language: req.Language,
		Code:     req.Code,
		Input:    req.Input,
	})
	return judge.RunResult{
		Output:    result.Output,
		Stderr:    result.Stderr,
		ExitCode:  result.ExitCode,
		TimedOut:  result.TimedOut,
		RuntimeMS: result.RuntimeMS,
	}, err
}
