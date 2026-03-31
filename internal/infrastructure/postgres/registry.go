package postgres

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/lib/pq"
	"github.com/tuannm99/judge-loop/internal/domain"
	outport "github.com/tuannm99/judge-loop/internal/port/out"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// RegistryStore handles registry version queries.
type RegistryStore struct{ db *DB }

// NewRegistryStore creates a new RegistryStore.
func NewRegistryStore(db *DB) *RegistryStore { return &RegistryStore{db: db} }

// GetLatest returns the most recently saved registry version, or nil if none.
func (s *RegistryStore) GetLatest(ctx context.Context) (*outport.RegistryVersion, error) {
	var model registryVersionModel
	err := s.db.Gorm.WithContext(ctx).Order("id DESC").Take(&model).Error
	if err == gorm.ErrRecordNotFound {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("get latest registry version: %w", err)
	}

	return &outport.RegistryVersion{
		Version:   model.Version,
		UpdatedAt: model.UpdatedAt,
		SyncedAt:  model.SyncedAt,
	}, nil
}

// Save inserts a new registry version record.
func (s *RegistryStore) Save(
	ctx context.Context,
	version string,
	updatedAt time.Time,
	refs []domain.ManifestRef,
) error {
	manifests, err := json.Marshal(refs)
	if err != nil {
		return fmt.Errorf("marshal manifests: %w", err)
	}

	model := registryVersionModel{
		Version:   version,
		UpdatedAt: updatedAt,
		Manifests: manifests,
		SyncedAt:  time.Now().UTC(),
	}
	if err := s.db.Gorm.WithContext(ctx).Create(&model).Error; err != nil {
		return fmt.Errorf("save registry version: %w", err)
	}
	return nil
}

// UpsertFromManifest inserts or updates a problem row from a ProblemManifest.
// The (provider, external_id) unique constraint is the conflict target.
func (s *ProblemStore) UpsertFromManifest(ctx context.Context, m domain.ProblemManifest) error {
	model := problemModel{
		Slug:          m.Slug,
		Title:         m.Title,
		Difficulty:    string(m.Difficulty),
		Tags:          pq.StringArray(m.Tags),
		PatternTags:   pq.StringArray(m.PatternTags),
		Provider:      string(m.Provider),
		ExternalID:    m.ExternalID,
		SourceURL:     m.SourceURL,
		EstimatedTime: m.EstimatedTime,
	}

	err := s.db.Gorm.WithContext(ctx).Clauses(clause.OnConflict{
		Columns: []clause.Column{{Name: "provider"}, {Name: "external_id"}},
		DoUpdates: clause.Assignments(map[string]any{
			"slug":           model.Slug,
			"title":          model.Title,
			"difficulty":     model.Difficulty,
			"tags":           model.Tags,
			"pattern_tags":   model.PatternTags,
			"source_url":     model.SourceURL,
			"estimated_time": model.EstimatedTime,
			"updated_at":     gorm.Expr("NOW()"),
		}),
	}).Create(&model).Error
	if err != nil {
		return fmt.Errorf("upsert problem %s/%s: %w", m.Provider, m.Slug, err)
	}
	return nil
}
