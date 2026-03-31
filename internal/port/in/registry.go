package in

import (
	"context"
	"time"

	"github.com/tuannm99/judge-loop/internal/domain"
	out "github.com/tuannm99/judge-loop/internal/port/out"
)

type RegistryService interface {
	SyncRegistry(
		ctx context.Context,
		version string,
		updatedAt time.Time,
		problems []domain.ProblemManifest,
		manifests []domain.ManifestRef,
	) (int, error)
	GetRegistryVersion(ctx context.Context) (*out.RegistryVersion, error)
}
