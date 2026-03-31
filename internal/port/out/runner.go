package out

import (
	"context"

	"github.com/tuannm99/judge-loop/internal/infrastructure/sandbox"
)

type CodeRunner interface {
	Run(ctx context.Context, req RunRequest) (sandbox.RunResult, error)
}
