package storage

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/tuannm99/judge-loop/internal/domain"
)

// RegistryStore handles registry version queries.
type RegistryStore struct{ db *DB }

// NewRegistryStore creates a new RegistryStore.
func NewRegistryStore(db *DB) *RegistryStore { return &RegistryStore{db: db} }

// RegistryVersionRow mirrors the registry_versions table.
type RegistryVersionRow struct {
	Version   string
	UpdatedAt time.Time
	SyncedAt  time.Time
}

// GetLatest returns the most recently saved registry version, or nil if none.
func (s *RegistryStore) GetLatest(ctx context.Context) (*RegistryVersionRow, error) {
	row := s.db.Pool.QueryRow(ctx, `
		SELECT version, updated_at, synced_at
		FROM registry_versions
		ORDER BY id DESC
		LIMIT 1`)

	var r RegistryVersionRow
	if err := row.Scan(&r.Version, &r.UpdatedAt, &r.SyncedAt); err != nil {
		return nil, nil // no rows yet — not an error
	}
	return &r, nil
}

// Save inserts a new registry version record.
func (s *RegistryStore) Save(ctx context.Context, version string, updatedAt time.Time, refs []domain.ManifestRef) error {
	manifests, err := json.Marshal(refs)
	if err != nil {
		return fmt.Errorf("marshal manifests: %w", err)
	}
	_, err = s.db.Pool.Exec(ctx, `
		INSERT INTO registry_versions (version, updated_at, manifests)
		VALUES ($1, $2, $3)`,
		version, updatedAt, manifests,
	)
	if err != nil {
		return fmt.Errorf("save registry version: %w", err)
	}
	return nil
}

// UpsertFromManifest inserts or updates a problem row from a ProblemManifest.
// The (provider, external_id) unique constraint is the conflict target.
func (s *ProblemStore) UpsertFromManifest(ctx context.Context, m domain.ProblemManifest) error {
	_, err := s.db.Pool.Exec(ctx, `
		INSERT INTO problems
			(slug, title, difficulty, tags, pattern_tags, provider, external_id, source_url, estimated_time)
		VALUES
			($1, $2, $3::difficulty, $4, $5, $6::provider, $7, $8, $9)
		ON CONFLICT (provider, external_id) DO UPDATE SET
			slug           = EXCLUDED.slug,
			title          = EXCLUDED.title,
			difficulty     = EXCLUDED.difficulty,
			tags           = EXCLUDED.tags,
			pattern_tags   = EXCLUDED.pattern_tags,
			source_url     = EXCLUDED.source_url,
			estimated_time = EXCLUDED.estimated_time,
			updated_at     = NOW()`,
		m.Slug, m.Title, string(m.Difficulty),
		m.Tags, m.PatternTags, string(m.Provider),
		m.ExternalID, m.SourceURL, m.EstimatedTime,
	)
	if err != nil {
		return fmt.Errorf("upsert problem %s/%s: %w", m.Provider, m.Slug, err)
	}
	return nil
}
