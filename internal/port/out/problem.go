package out

import (
	"context"

	"github.com/google/uuid"
	"github.com/tuannm99/judge-loop/internal/domain"
)

type ProblemRepository interface {
	List(ctx context.Context, filter ProblemFilter) ([]domain.Problem, int, error)
	GetByID(ctx context.Context, id uuid.UUID) (*domain.Problem, error)
	GetBySlug(ctx context.Context, slug string) (*domain.Problem, error)
	Suggest(ctx context.Context, userID uuid.UUID, patterns []string) (*domain.Problem, error)
	GetUnsolved(ctx context.Context, userID uuid.UUID, limit int) ([]domain.Problem, error)
	UpsertFromManifest(ctx context.Context, m domain.ProblemManifest) error
}
