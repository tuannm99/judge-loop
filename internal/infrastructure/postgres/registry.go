package postgres

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/tuannm99/judge-loop/internal/domain"
	outport "github.com/tuannm99/judge-loop/internal/port/out"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// RegistryRepositoryImpl handles registry version queries.
type RegistryRepositoryImpl struct{ db *DB }

var _ outport.RegistryRepository = (*RegistryRepositoryImpl)(nil)

// NewRegistryRepositoryImpl creates a new RegistryRepositoryImpl.
func NewRegistryRepositoryImpl(db *DB) *RegistryRepositoryImpl {
	return &RegistryRepositoryImpl{db: db}
}

// GetLatest returns the most recently saved registry version, or nil if none.
func (s *RegistryRepositoryImpl) GetLatest(ctx context.Context) (*outport.RegistryVersion, error) {
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
func (s *RegistryRepositoryImpl) Save(
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
func (s *ProblemRepositoryImpl) UpsertFromManifest(ctx context.Context, m domain.ProblemManifest) error {
	starterCode, err := json.Marshal(m.StarterCode)
	if err != nil {
		return fmt.Errorf("marshal starter code: %w", err)
	}

	model := problemModel{
		Slug:          m.Slug,
		Title:         m.Title,
		Difficulty:    string(m.Difficulty),
		Provider:      string(m.Provider),
		ExternalID:    m.ExternalID,
		SourceURL:     m.SourceURL,
		EstimatedTime: m.EstimatedTime,
		StarterCode:   starterCode,
	}

	err = s.db.Gorm.WithContext(ctx).Clauses(clause.OnConflict{
		Columns: []clause.Column{{Name: "provider"}, {Name: "external_id"}},
		DoUpdates: clause.Assignments(map[string]any{
			"slug":           model.Slug,
			"title":          model.Title,
			"difficulty":     model.Difficulty,
			"source_url":     model.SourceURL,
			"estimated_time": model.EstimatedTime,
			"starter_code":   model.StarterCode,
			"updated_at":     gorm.Expr("NOW()"),
		}),
	}).Create(&model).Error
	if err != nil {
		return fmt.Errorf("upsert problem %s/%s: %w", m.Provider, m.Slug, err)
	}

	var saved problemModel
	if err := s.db.Gorm.WithContext(ctx).
		Where("provider = ? AND external_id = ?", model.Provider, model.ExternalID).
		Take(&saved).Error; err != nil {
		return fmt.Errorf("load upserted problem %s/%s: %w", m.Provider, m.Slug, err)
	}

	if err := s.replaceProblemLabels(ctx, saved.ID, "tag", m.Tags); err != nil {
		return err
	}
	if err := s.replaceProblemLabels(ctx, saved.ID, "pattern", m.PatternTags); err != nil {
		return err
	}
	return nil
}

func (s *ProblemRepositoryImpl) replaceProblemLabels(
	ctx context.Context,
	problemID uuid.UUID,
	kind string,
	slugs []string,
) error {
	return s.db.Gorm.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		subquery := tx.Table("problem_labels").Select("id").Where("kind = ?", kind)
		if err := tx.Where("problem_id = ? AND problem_label_id IN (?)", problemID, subquery).
			Delete(&problemLabelLinkModel{}).Error; err != nil {
			return fmt.Errorf("delete %s labels for problem %s: %w", kind, problemID, err)
		}

		for _, slug := range slugs {
			if slug == "" {
				continue
			}
			label := problemLabelModel{
				Kind: kind,
				Slug: slug,
				Name: slug,
			}
			if err := tx.Clauses(clause.OnConflict{
				Columns:   []clause.Column{{Name: "kind"}, {Name: "slug"}},
				DoUpdates: clause.Assignments(map[string]any{"name": slug, "updated_at": gorm.Expr("NOW()")}),
			}).Create(&label).Error; err != nil {
				return fmt.Errorf("upsert %s label %q: %w", kind, slug, err)
			}

			if err := tx.Where("kind = ? AND slug = ?", kind, slug).Take(&label).Error; err != nil {
				return fmt.Errorf("load %s label %q: %w", kind, slug, err)
			}

			link := problemLabelLinkModel{
				ProblemID:      problemID,
				ProblemLabelID: label.ID,
			}
			if err := tx.Clauses(clause.OnConflict{DoNothing: true}).Create(&link).Error; err != nil {
				return fmt.Errorf("link %s label %q: %w", kind, slug, err)
			}
		}
		return nil
	})
}
