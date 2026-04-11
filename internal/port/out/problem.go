package out

import (
	"context"

	"github.com/google/uuid"
	"github.com/tuannm99/judge-loop/internal/domain"
)

type ProblemRepository interface {
	List(ctx context.Context, filter ProblemFilter) ([]domain.Problem, int, error)
	ListLabels(ctx context.Context, kind string) ([]string, error)
	ListLabelRecords(ctx context.Context, kind string) ([]domain.ProblemLabel, error)
	CreateLabel(ctx context.Context, label domain.ProblemLabel) (*domain.ProblemLabel, error)
	UpdateLabel(ctx context.Context, label domain.ProblemLabel) (*domain.ProblemLabel, error)
	DeleteLabel(ctx context.Context, id uuid.UUID) error
	GetByID(ctx context.Context, id uuid.UUID) (*domain.Problem, error)
	GetBySlug(ctx context.Context, slug string) (*domain.Problem, error)
	Update(ctx context.Context, id uuid.UUID, m domain.ProblemManifest) (*domain.Problem, error)
	Suggest(ctx context.Context, userID uuid.UUID, patterns []string) (*domain.Problem, error)
	GetUnsolved(ctx context.Context, userID uuid.UUID, limit int) ([]domain.Problem, error)
	UpsertFromManifest(ctx context.Context, m domain.ProblemManifest) error
}
