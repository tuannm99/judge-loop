// Package registry loads and parses local problem registry manifests from disk.
package registry

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/tuannm99/judge-loop/internal/domain"
)

// Index is the top-level registry file (index.json).
type Index struct {
	Version   string               `json:"version"`
	UpdatedAt time.Time            `json:"updated_at"`
	Manifests []domain.ManifestRef `json:"manifests"`
}

// ProviderManifest is the parsed content of a providers/*.json file.
type ProviderManifest struct {
	Problems []domain.ProblemManifest `json:"problems"`
}

// TrackManifest is the parsed content of a tracks/*.json file.
type TrackManifest struct {
	Name        string                `json:"name"`
	Title       string                `json:"title"`
	Description string                `json:"description"`
	Problems    []domain.TrackProblem `json:"problems"`
}

// LoadIndex reads and parses the index.json from the registry directory.
func LoadIndex(registryPath string) (*Index, error) {
	data, err := os.ReadFile(filepath.Join(registryPath, "index.json"))
	if err != nil {
		return nil, fmt.Errorf("read index.json: %w", err)
	}
	var idx Index
	if err := json.Unmarshal(data, &idx); err != nil {
		return nil, fmt.Errorf("parse index.json: %w", err)
	}
	return &idx, nil
}

// LoadAllProblems reads every provider manifest listed in the index and
// returns the combined list of ProblemManifest entries.
// Track manifests are skipped — they only define ordering, not problem metadata.
func LoadAllProblems(registryPath string, idx *Index) ([]domain.ProblemManifest, error) {
	var out []domain.ProblemManifest
	for _, ref := range idx.Manifests {
		dir := filepath.Dir(ref.Path)
		if dir != "providers" {
			continue // tracks and roadmaps don't contain problem metadata
		}
		pm, err := loadProviderManifest(filepath.Join(registryPath, ref.Path))
		if err != nil {
			return nil, fmt.Errorf("load manifest %q: %w", ref.Name, err)
		}
		out = append(out, pm.Problems...)
	}
	return out, nil
}

func loadProviderManifest(path string) (*ProviderManifest, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var pm ProviderManifest
	if err := json.Unmarshal(data, &pm); err != nil {
		return nil, err
	}
	return &pm, nil
}
