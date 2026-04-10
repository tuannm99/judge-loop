package out

import (
	"context"

	"github.com/tuannm99/judge-loop/internal/domain/judge"
)

type CodeRunner interface {
	Run(ctx context.Context, req RunRequest) (judge.RunResult, error)
}
