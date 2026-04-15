package registry

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"github.com/tuannm99/judge-loop/internal/domain"
)

func TestLoadIndexAndProblems(t *testing.T) {
	dir := t.TempDir()
	require.NoError(t, os.MkdirAll(filepath.Join(dir, "providers", "leet", "free"), 0o755))
	require.NoError(t, os.Mkdir(filepath.Join(dir, "tracks"), 0o755))

	index := `{"version":"v1","updated_at":"2026-03-31T00:00:00Z","manifests":[{"name":"leet","path":"providers/leet/free/problems.json"},{"name":"roadmap","path":"tracks/base.json"}]}`
	provider := `{"problems":[{"slug":"two-sum","title":"Two Sum","difficulty":"easy","provider":"leetcode","external_id":"1","tags":["array"],"pattern_tags":["hash-map"]}]}`
	track := `{"name":"base","title":"Base","description":"x","problems":[{"slug":"two-sum","provider":"leetcode"}]}`
	require.NoError(t, os.WriteFile(filepath.Join(dir, "index.json"), []byte(index), 0o644))
	require.NoError(t, os.WriteFile(filepath.Join(dir, "providers", "leet", "free", "problems.json"), []byte(provider), 0o644))
	require.NoError(t, os.WriteFile(filepath.Join(dir, "tracks", "base.json"), []byte(track), 0o644))

	idx, err := LoadIndex(dir)
	require.NoError(t, err)
	require.Equal(t, "v1", idx.Version)
	require.Equal(t, time.Date(2026, 3, 31, 0, 0, 0, 0, time.UTC), idx.UpdatedAt)

	problems, err := LoadAllProblems(dir, idx)
	require.NoError(t, err)
	require.Len(t, problems, 1)
	require.Equal(t, "two-sum", problems[0].Slug)
	require.ElementsMatch(t, []string{"array", "hash-map"}, problems[0].Tags)

	pm, err := loadProviderManifest(filepath.Join(dir, "providers", "leet", "free", "problems.json"))
	require.NoError(t, err)
	require.Len(t, pm.Problems, 1)
	require.ElementsMatch(t, []string{"array", "hash-map"}, pm.Problems[0].Tags)
}

func TestLoadIndexErrors(t *testing.T) {
	_, err := LoadIndex(t.TempDir())
	require.Error(t, err)

	dir := t.TempDir()
	require.NoError(t, os.WriteFile(filepath.Join(dir, "index.json"), []byte("{"), 0o644))
	_, err = LoadIndex(dir)
	require.Error(t, err)

	_, err = loadProviderManifest(filepath.Join(dir, "missing.json"))
	require.Error(t, err)
}

func TestLoadAllProblemsReturnsManifestError(t *testing.T) {
	dir := t.TempDir()
	require.NoError(t, os.Mkdir(filepath.Join(dir, "providers"), 0o755))
	require.NoError(t, os.WriteFile(filepath.Join(dir, "providers", "bad.json"), []byte("{"), 0o644))

	_, err := LoadAllProblems(dir, &Index{
		Manifests: []domain.ManifestRef{{Name: "bad", Path: "providers/bad.json"}},
	})
	require.Error(t, err)
}
