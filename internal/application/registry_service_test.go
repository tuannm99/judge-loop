package application

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"github.com/tuannm99/judge-loop/internal/domain"
	outport "github.com/tuannm99/judge-loop/internal/port/out"
	outmocks "github.com/tuannm99/judge-loop/internal/port/out/mocks"
)

func TestRegistryServiceGetRegistryVersion(t *testing.T) {
	problems := outmocks.NewMockProblemRepository(t)
	registry := outmocks.NewMockRegistryRepository(t)
	service := NewRegistryService(problems, registry)

	ctx := context.Background()
	version := &outport.RegistryVersion{Version: "v1"}
	registry.EXPECT().GetLatest(ctx).Return(version, nil)

	got, err := service.GetRegistryVersion(ctx)
	require.NoError(t, err)
	require.Equal(t, version, got)
}

func TestRegistryServiceSyncRegistrySuccess(t *testing.T) {
	problems := outmocks.NewMockProblemRepository(t)
	registry := outmocks.NewMockRegistryRepository(t)
	service := NewRegistryService(problems, registry)

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

func TestRegistryServiceErrors(t *testing.T) {
	problems := outmocks.NewMockProblemRepository(t)
	registry := outmocks.NewMockRegistryRepository(t)

	ctx := context.Background()
	manifest := domain.ProblemManifest{Slug: "two-sum"}

	problems.EXPECT().UpsertFromManifest(ctx, manifest).Return(errors.New("upsert"))
	_, err := NewRegistryService(problems, registry).SyncRegistry(ctx, "v1", time.Time{}, []domain.ProblemManifest{manifest}, nil)
	require.Error(t, err)

	problems2 := outmocks.NewMockProblemRepository(t)
	registry2 := outmocks.NewMockRegistryRepository(t)
	problems2.EXPECT().UpsertFromManifest(ctx, manifest).Return(nil)
	registry2.EXPECT().Save(ctx, "v1", mock.Anything, []domain.ManifestRef(nil)).Return(errors.New("save"))
	_, err = NewRegistryService(problems2, registry2).SyncRegistry(ctx, "v1", time.Time{}, []domain.ProblemManifest{manifest}, nil)
	require.Error(t, err)
}
