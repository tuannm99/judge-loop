package application

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/tuannm99/judge-loop/internal/domain"
	inport "github.com/tuannm99/judge-loop/internal/port/in"
	outport "github.com/tuannm99/judge-loop/internal/port/out"
)

type RegistryService struct {
	problems  outport.ProblemRepository
	testCases outport.TestCaseRepository
	registry  outport.RegistryRepository
}

var _ inport.RegistryService = (*RegistryService)(nil)

func NewRegistryService(
	problems outport.ProblemRepository,
	testCases outport.TestCaseRepository,
	registry outport.RegistryRepository,
) *RegistryService {
	return &RegistryService{
		problems:  problems,
		testCases: testCases,
		registry:  registry,
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
		if len(m.TestCases) > 0 && s.testCases != nil {
			problem, err := s.problems.GetBySlug(ctx, m.Slug)
			if err != nil {
				return 0, err
			}
			if problem == nil {
				return 0, fmt.Errorf("problem not found after registry upsert: %s", m.Slug)
			}
			if err := s.testCases.ReplaceForProblem(ctx, problem.ID, testCaseManifestsToDomain(problem.ID, m.TestCases)); err != nil {
				return 0, err
			}
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

func testCaseManifestsToDomain(problemID uuid.UUID, manifests []domain.TestCaseManifest) []domain.TestCase {
	out := make([]domain.TestCase, 0, len(manifests))
	for i, manifest := range manifests {
		out = append(out, domain.TestCase{
			ProblemID: problemID,
			Input:     manifest.Input,
			Expected:  manifest.Expected,
			IsHidden:  manifest.IsHidden,
			OrderIdx:  i,
		})
	}
	return out
}

func (s *RegistryService) GetRegistryVersion(ctx context.Context) (*outport.RegistryVersion, error) {
	return s.registry.GetLatest(ctx)
}
