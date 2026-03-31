package application

import (
	"context"
	"time"

	"github.com/tuannm99/judge-loop/internal/domain"
	inport "github.com/tuannm99/judge-loop/internal/port/in"
	outport "github.com/tuannm99/judge-loop/internal/port/out"
)

type RegistryService struct {
	problems outport.ProblemRepository
	registry outport.RegistryRepository
}

var _ inport.RegistryService = (*RegistryService)(nil)

func NewRegistryService(
	problems outport.ProblemRepository,
	registry outport.RegistryRepository,
) *RegistryService {
	return &RegistryService{
		problems: problems,
		registry: registry,
	}
}

func (s *RegistryService) SyncRegistry(
	ctx context.Context,
	version string,
	updatedAt time.Time,
	problems []domain.ProblemManifest,
	manifests []domain.ManifestRef,
) (int, error) {
	synced := 0
	for _, m := range problems {
		if err := s.problems.UpsertFromManifest(ctx, m); err != nil {
			return 0, err
		}
		synced++
	}
	if updatedAt.IsZero() {
		updatedAt = time.Now()
	}
	if err := s.registry.Save(ctx, version, updatedAt, manifests); err != nil {
		return 0, err
	}
	return synced, nil
}

func (s *RegistryService) GetRegistryVersion(ctx context.Context) (*outport.RegistryVersion, error) {
	return s.registry.GetLatest(ctx)
}
