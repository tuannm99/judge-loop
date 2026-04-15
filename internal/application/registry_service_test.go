package application

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"github.com/tuannm99/judge-loop/internal/domain"
	outport "github.com/tuannm99/judge-loop/internal/port/out"
	outmocks "github.com/tuannm99/judge-loop/internal/port/out/mocks"
)

func TestRegistryServiceGetRegistryVersion(t *testing.T) {
	problems := outmocks.NewMockProblemRepository(t)
	testCases := outmocks.NewMockTestCaseRepository(t)
	registry := outmocks.NewMockRegistryRepository(t)
	service := NewRegistryService(problems, testCases, registry)

	ctx := context.Background()
	version := &outport.RegistryVersion{Version: "v1"}
	registry.EXPECT().GetLatest(ctx).Return(version, nil)

	got, err := service.GetRegistryVersion(ctx)
	require.NoError(t, err)
	require.Equal(t, version, got)
}

func TestRegistryServiceSyncRegistrySuccess(t *testing.T) {
	problems := outmocks.NewMockProblemRepository(t)
	testCases := outmocks.NewMockTestCaseRepository(t)
	registry := outmocks.NewMockRegistryRepository(t)
	service := NewRegistryService(problems, testCases, registry)

	ctx := context.Background()
	updatedAt := time.Date(2026, 3, 31, 10, 0, 0, 0, time.UTC)
	problemsList := []domain.ProblemManifest{{Slug: "a"}, {Slug: "b"}}
	refs := []domain.ManifestRef{{Name: "main", Path: "providers/main.json"}}
	for _, problem := range problemsList {
		problems.EXPECT().UpsertFromManifest(ctx, problem).Return(nil)
	}
	registry.EXPECT().Save(ctx, "v1", updatedAt, refs).Return(nil)

	synced, err := service.SyncRegistry(ctx, "v1", updatedAt, problemsList, refs)
	require.NoError(t, err)
	require.Equal(t, 2, synced)
}

func TestRegistryServiceSyncRegistryImportsTestCases(t *testing.T) {
	problems := outmocks.NewMockProblemRepository(t)
	testCases := outmocks.NewMockTestCaseRepository(t)
	registry := outmocks.NewMockRegistryRepository(t)
	service := NewRegistryService(problems, testCases, registry)

	ctx := context.Background()
	problemID := uuid.New()
	manifest := domain.ProblemManifest{
		Slug: "two-sum",
		TestCases: []domain.TestCaseManifest{
			{Input: `{"nums":[2,7,11,15],"target":9}`, Expected: `[0,1]`},
			{Input: `{"nums":[3,2,4],"target":6}`, Expected: `[1,2]`, IsHidden: true},
		},
	}

	problems.EXPECT().UpsertFromManifest(ctx, manifest).Return(nil)
	problems.EXPECT().GetBySlug(ctx, "two-sum").Return(&domain.Problem{ID: problemID}, nil)
	testCases.EXPECT().
		ReplaceForProblem(ctx, problemID, mock.MatchedBy(func(cases []domain.TestCase) bool {
			return len(cases) == 2 &&
				cases[0].ProblemID == problemID &&
				cases[0].Input == manifest.TestCases[0].Input &&
				cases[0].Expected == manifest.TestCases[0].Expected &&
				!cases[0].IsHidden &&
				cases[0].OrderIdx == 0 &&
				cases[1].IsHidden &&
				cases[1].OrderIdx == 1
		})).
		Return(nil)
	registry.EXPECT().Save(ctx, "v1", mock.Anything, []domain.ManifestRef(nil)).Return(nil)

	synced, err := service.SyncRegistry(ctx, "v1", time.Time{}, []domain.ProblemManifest{manifest}, nil)
	require.NoError(t, err)
	require.Equal(t, 1, synced)
}

func TestRegistryServiceErrors(t *testing.T) {
	problems := outmocks.NewMockProblemRepository(t)
	testCases := outmocks.NewMockTestCaseRepository(t)
	registry := outmocks.NewMockRegistryRepository(t)

	ctx := context.Background()
	manifest := domain.ProblemManifest{Slug: "two-sum"}

	problems.EXPECT().UpsertFromManifest(ctx, manifest).Return(errors.New("upsert"))
	_, err := NewRegistryService(problems, testCases, registry).SyncRegistry(ctx, "v1", time.Time{}, []domain.ProblemManifest{manifest}, nil)
	require.Error(t, err)

	problems2 := outmocks.NewMockProblemRepository(t)
	testCases2 := outmocks.NewMockTestCaseRepository(t)
	registry2 := outmocks.NewMockRegistryRepository(t)
	problems2.EXPECT().UpsertFromManifest(ctx, manifest).Return(nil)
	registry2.EXPECT().Save(ctx, "v1", mock.Anything, []domain.ManifestRef(nil)).Return(errors.New("save"))
	_, err = NewRegistryService(problems2, testCases2, registry2).SyncRegistry(ctx, "v1", time.Time{}, []domain.ProblemManifest{manifest}, nil)
	require.Error(t, err)
}
