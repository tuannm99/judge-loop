package postgres

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/require"
	"github.com/tuannm99/judge-loop/internal/domain"
)

func TestRegistryRepositoryGetLatestAndSave(t *testing.T) {
	db, mock := newMockDB(t)
	repository := NewRegistryRepositoryImpl(db)
	updatedAt := time.Date(2026, 6, 27, 1, 0, 0, 0, time.UTC)
	syncedAt := updatedAt.Add(time.Minute)

	mock.ExpectQuery(`SELECT \* FROM "registry_versions" ORDER BY id DESC LIMIT \$1`).
		WithArgs(1).
		WillReturnRows(sqlmock.NewRows([]string{
			"id", "version", "updated_at", "manifests", "synced_at",
		}).AddRow(1, "v1", updatedAt, []byte(`[]`), syncedAt))

	version, err := repository.GetLatest(context.Background())
	require.NoError(t, err)
	require.Equal(t, "v1", version.Version)
	require.Equal(t, syncedAt, version.SyncedAt)

	mock.ExpectQuery(`INSERT INTO "registry_versions".*RETURNING "id"`).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(2))

	err = repository.Save(context.Background(), "v2", updatedAt, []domain.ManifestRef{{
		Name: "leetcode",
		Path: "providers/leetcode/free/problems.json",
	}})
	require.NoError(t, err)
}

func TestRegistryRepositoryGetLatestHandlesNotFoundAndError(t *testing.T) {
	t.Run("not found", func(t *testing.T) {
		db, mock := newMockDB(t)
		repository := NewRegistryRepositoryImpl(db)
		mock.ExpectQuery(`SELECT \* FROM "registry_versions"`).
			WillReturnRows(sqlmock.NewRows([]string{"id"}))

		version, err := repository.GetLatest(context.Background())
		require.NoError(t, err)
		require.Nil(t, version)
	})

	t.Run("query error", func(t *testing.T) {
		db, mock := newMockDB(t)
		repository := NewRegistryRepositoryImpl(db)
		mock.ExpectQuery(`SELECT \* FROM "registry_versions"`).
			WillReturnError(errors.New("database unavailable"))

		version, err := repository.GetLatest(context.Background())
		require.Nil(t, version)
		require.ErrorContains(t, err, "get latest registry version: database unavailable")
	})
}
