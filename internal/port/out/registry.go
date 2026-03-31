package out

import (
	"context"
	"time"

	"github.com/tuannm99/judge-loop/internal/domain"
)

type RegistryRepository interface {
	GetLatest(ctx context.Context) (*RegistryVersion, error)
	Save(ctx context.Context, version string, updatedAt time.Time, refs []domain.ManifestRef) error
}
