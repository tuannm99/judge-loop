package postgres

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
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
	if m.StarterCode == nil {
		m.StarterCode = map[string]string{}
	}

	starterCode, err := json.Marshal(m.StarterCode)
	if err != nil {
		return fmt.Errorf("marshal starter code: %w", err)
	}

	model := problemModel{
		Slug:                m.Slug,
		Title:               m.Title,
		Difficulty:          string(m.Difficulty),
		Provider:            string(m.Provider),
		ExternalID:          m.ExternalID,
		SourceURL:           m.SourceURL,
		EstimatedTime:       m.EstimatedTime,
		DescriptionMarkdown: m.DescriptionMarkdown,
		StarterCode:         starterCode,
	}

	err = s.db.Gorm.WithContext(ctx).Clauses(clause.OnConflict{
		Columns: []clause.Column{{Name: "provider"}, {Name: "external_id"}},
		DoUpdates: clause.Assignments(map[string]any{
			"slug":                 model.Slug,
			"title":                model.Title,
			"difficulty":           model.Difficulty,
			"source_url":           model.SourceURL,
			"estimated_time":       model.EstimatedTime,
			"description_markdown": model.DescriptionMarkdown,
			"starter_code":         model.StarterCode,
			"updated_at":           gorm.Expr("NOW()"),
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

	if err := s.replaceProblemLabels(ctx, saved.ID, m.Tags); err != nil {
		return err
	}
	return nil
}

func (s *ProblemRepositoryImpl) replaceProblemLabels(
	ctx context.Context,
	problemID uuid.UUID,
	slugs []string,
) error {
	return s.db.Gorm.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		return replaceProblemLabelsWithTx(tx, problemID, slugs)
	})
}

func replaceProblemLabelsWithTx(
	tx *gorm.DB,
	problemID uuid.UUID,
	slugs []string,
) error {
	normalized := make([]string, 0, len(slugs))
	seen := make(map[string]struct{}, len(slugs))
	for _, slug := range slugs {
		slug = strings.TrimSpace(slug)
		if slug == "" {
			continue
		}
		if _, ok := seen[slug]; ok {
			continue
		}
		seen[slug] = struct{}{}
		normalized = append(normalized, slug)
	}

	return tx.Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("problem_id = ?", problemID).
			Delete(&problemLabelLinkModel{}).Error; err != nil {
			return fmt.Errorf("delete labels for problem %s: %w", problemID, err)
		}

		for _, slug := range normalized {
			label := problemLabelModel{
				Slug: slug,
				Name: slug,
			}
			if err := tx.Clauses(clause.OnConflict{
				Columns:   []clause.Column{{Name: "slug"}},
				DoUpdates: clause.Assignments(map[string]any{"name": slug, "updated_at": gorm.Expr("NOW()")}),
			}).Create(&label).Error; err != nil {
				return fmt.Errorf("upsert label %q: %w", slug, err)
			}

			if err := tx.Where("slug = ?", slug).Take(&label).Error; err != nil {
				return fmt.Errorf("load label %q: %w", slug, err)
			}

			link := problemLabelLinkModel{
				ProblemID:      problemID,
				ProblemLabelID: label.ID,
			}
			if err := tx.Clauses(clause.OnConflict{DoNothing: true}).Create(&link).Error; err != nil {
				return fmt.Errorf("link label %q: %w", slug, err)
			}
		}
		return nil
	})
}
