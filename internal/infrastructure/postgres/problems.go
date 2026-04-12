package postgres

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/google/uuid"
	"github.com/tuannm99/judge-loop/internal/domain"
	outport "github.com/tuannm99/judge-loop/internal/port/out"
	"gorm.io/gorm"
)

// ProblemRepositoryImpl handles all problem queries.
type ProblemRepositoryImpl struct{ db *DB }

var _ outport.ProblemRepository = (*ProblemRepositoryImpl)(nil)

// NewProblemRepositoryImpl creates a new ProblemRepositoryImpl.
func NewProblemRepositoryImpl(db *DB) *ProblemRepositoryImpl { return &ProblemRepositoryImpl{db: db} }

// ProblemFilter holds optional filters for listing problems.
// Nil pointer fields mean "no filter".
type ProblemFilter = outport.ProblemFilter

// List returns problems matching the filter with a total count.
func (s *ProblemRepositoryImpl) List(ctx context.Context, f ProblemFilter) ([]domain.Problem, int, error) {
	if f.Limit <= 0 {
		f.Limit = 20
	}

	baseQuery := s.db.Gorm.WithContext(ctx).Model(&problemModel{})
	if f.Difficulty != nil {
		baseQuery = baseQuery.Where("difficulty = ?", string(*f.Difficulty))
	}
	if f.Provider != nil {
		baseQuery = baseQuery.Where("provider = ?", string(*f.Provider))
	}
	if len(f.Tags) > 0 {
		baseQuery = joinProblemLabelFilter(baseQuery, f.Tags)
	}

	var total int64
	if err := baseQuery.Session(&gorm.Session{}).Distinct("problems.id").Count(&total).Error; err != nil {
		return nil, 0, fmt.Errorf("count problems: %w", err)
	}

	var models []problemModel
	if err := baseQuery.Session(&gorm.Session{}).
		Distinct("problems.*").
		Order("created_at DESC").
		Limit(f.Limit).
		Offset(f.Offset).
		Find(&models).Error; err != nil {
		return nil, 0, fmt.Errorf("list problems: %w", err)
	}
	modelPtrs := make([]*problemModel, 0, len(models))
	for i := range models {
		modelPtrs = append(modelPtrs, &models[i])
	}
	if err := s.loadProblemLabels(ctx, modelPtrs); err != nil {
		return nil, 0, err
	}

	out := make([]domain.Problem, 0, len(models))
	for _, model := range models {
		out = append(out, model.toDomain())
	}
	return out, int(total), nil
}

func (s *ProblemRepositoryImpl) ListLabels(ctx context.Context, kind string) ([]string, error) {
	var labels []string
	if err := s.db.Gorm.WithContext(ctx).
		Table("problem_labels").
		Order("slug ASC").
		Pluck("slug", &labels).Error; err != nil {
		return nil, fmt.Errorf("list %s labels: %w", kind, err)
	}
	return labels, nil
}

func (s *ProblemRepositoryImpl) ListLabelRecords(ctx context.Context, kind string) ([]domain.ProblemLabel, error) {
	var models []problemLabelModel
	if err := s.db.Gorm.WithContext(ctx).
		Order("name ASC, slug ASC").
		Find(&models).Error; err != nil {
		return nil, fmt.Errorf("list %s label records: %w", kind, err)
	}

	labels := make([]domain.ProblemLabel, 0, len(models))
	for _, model := range models {
		labels = append(labels, model.toDomain())
	}
	return labels, nil
}

func (s *ProblemRepositoryImpl) CreateLabel(ctx context.Context, label domain.ProblemLabel) (*domain.ProblemLabel, error) {
	model := problemLabelModel{
		Slug: label.Slug,
		Name: label.Name,
	}
	if err := s.db.Gorm.WithContext(ctx).Create(&model).Error; err != nil {
		return nil, fmt.Errorf("create label %s/%s: %w", label.Kind, label.Slug, err)
	}
	out := model.toDomain()
	return &out, nil
}

func (s *ProblemRepositoryImpl) UpdateLabel(ctx context.Context, label domain.ProblemLabel) (*domain.ProblemLabel, error) {
	updates := map[string]any{
		"slug":       label.Slug,
		"name":       label.Name,
		"updated_at": gorm.Expr("NOW()"),
	}
	if err := s.db.Gorm.WithContext(ctx).
		Model(&problemLabelModel{}).
		Where("id = ?", label.ID).
		Updates(updates).Error; err != nil {
		return nil, fmt.Errorf("update label %s: %w", label.ID, err)
	}

	var model problemLabelModel
	if err := s.db.Gorm.WithContext(ctx).First(&model, "id = ?", label.ID).Error; err != nil {
		return nil, fmt.Errorf("load updated label %s: %w", label.ID, err)
	}
	out := model.toDomain()
	return &out, nil
}

func (s *ProblemRepositoryImpl) DeleteLabel(ctx context.Context, id uuid.UUID) error {
	if err := s.db.Gorm.WithContext(ctx).Delete(&problemLabelModel{}, "id = ?", id).Error; err != nil {
		return fmt.Errorf("delete label %s: %w", id, err)
	}
	return nil
}

// GetByID returns a problem by UUID, or nil if not found.
func (s *ProblemRepositoryImpl) GetByID(ctx context.Context, id uuid.UUID) (*domain.Problem, error) {
	var model problemModel
	err := s.db.Gorm.WithContext(ctx).First(&model, "id = ?", id).Error
	if err == gorm.ErrRecordNotFound {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("get problem by id: %w", err)
	}
	if err := s.loadProblemLabels(ctx, []*problemModel{&model}); err != nil {
		return nil, err
	}
	p := model.toDomain()
	return &p, nil
}

// GetBySlug returns a problem by slug, or nil if not found.
func (s *ProblemRepositoryImpl) GetBySlug(ctx context.Context, slug string) (*domain.Problem, error) {
	var model problemModel
	err := s.db.Gorm.WithContext(ctx).First(&model, "slug = ?", slug).Error
	if err == gorm.ErrRecordNotFound {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("get problem by slug: %w", err)
	}
	if err := s.loadProblemLabels(ctx, []*problemModel{&model}); err != nil {
		return nil, err
	}
	p := model.toDomain()
	return &p, nil
}

func (s *ProblemRepositoryImpl) Update(
	ctx context.Context,
	id uuid.UUID,
	m domain.ProblemManifest,
) (*domain.Problem, error) {
	if m.StarterCode == nil {
		m.StarterCode = map[string]string{}
	}

	starterCode, err := json.Marshal(m.StarterCode)
	if err != nil {
		return nil, fmt.Errorf("marshal starter code: %w", err)
	}

	err = s.db.Gorm.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		updates := map[string]any{
			"slug":           m.Slug,
			"title":          m.Title,
			"difficulty":     string(m.Difficulty),
			"provider":       string(m.Provider),
			"external_id":    m.ExternalID,
			"source_url":     m.SourceURL,
			"estimated_time": m.EstimatedTime,
			"starter_code":   starterCode,
			"updated_at":     gorm.Expr("NOW()"),
		}
		result := tx.Model(&problemModel{}).Where("id = ?", id).Updates(updates)
		if result.Error != nil {
			return fmt.Errorf("update problem %s: %w", id, result.Error)
		}
		if result.RowsAffected == 0 {
			return nil
		}
		if err := replaceProblemLabelsWithTx(tx, id, m.Tags); err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		return nil, err
	}

	return s.GetByID(ctx, id)
}

// Suggest returns a random unsolved problem for the user.
// If tags is non-empty, problems matching those tags are prioritized.
func (s *ProblemRepositoryImpl) Suggest(
	ctx context.Context,
	userID uuid.UUID,
	tags []string,
) (*domain.Problem, error) {
	if tags == nil {
		tags = []string{}
	}

	q := s.db.Gorm.WithContext(ctx).
		Model(&problemModel{}).
		Where("problems.id NOT IN (?)",
			s.db.Gorm.WithContext(ctx).
				Model(&submissionModel{}).
				Select("DISTINCT problem_id").
				Where("user_id = ? AND status = ?", userID, string(domain.StatusAccepted)),
		)

	if len(tags) > 0 {
		q = q.Order(gorm.Expr(`
				CASE WHEN EXISTS (
					SELECT 1
					FROM problem_label_links pll
					JOIN problem_labels pl ON pl.id = pll.problem_label_id
					WHERE pll.problem_id = problems.id
					  AND pl.slug IN ?
				) THEN 0 ELSE 1 END
			`, tags))
	}

	var model problemModel
	err := q.Order("RANDOM()").Limit(1).Take(&model).Error
	if err == gorm.ErrRecordNotFound {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("suggest problem: %w", err)
	}
	if err := s.loadProblemLabels(ctx, []*problemModel{&model}); err != nil {
		return nil, err
	}
	p := model.toDomain()
	return &p, nil
}

func joinProblemLabelFilter(q *gorm.DB, slugs []string) *gorm.DB {
	linkAlias := "tag_pll"
	labelAlias := "tag_pl"

	return q.Joins(fmt.Sprintf(
		"JOIN problem_label_links %s ON %s.problem_id = problems.id",
		linkAlias, linkAlias,
	)).Joins(fmt.Sprintf(
		"JOIN problem_labels %s ON %s.id = %s.problem_label_id AND %s.slug IN ?",
		labelAlias, labelAlias, linkAlias, labelAlias,
	), slugs)
}

func (s *ProblemRepositoryImpl) loadProblemLabels(ctx context.Context, models []*problemModel) error {
	if len(models) == 0 {
		return nil
	}

	ids := make([]uuid.UUID, 0, len(models))
	indexByID := make(map[uuid.UUID]*problemModel, len(models))
	for _, model := range models {
		ids = append(ids, model.ID)
		indexByID[model.ID] = model
	}

	type row struct {
		ProblemID uuid.UUID
		Slug      string
	}

	var rows []row
	if err := s.db.Gorm.WithContext(ctx).
		Table("problem_label_links pll").
		Select("pll.problem_id, pl.slug").
		Joins("JOIN problem_labels pl ON pl.id = pll.problem_label_id").
		Where("pll.problem_id IN ?", ids).
		Order("pl.slug ASC").
		Scan(&rows).Error; err != nil {
		return fmt.Errorf("load problem labels: %w", err)
	}

	for _, row := range rows {
		model := indexByID[row.ProblemID]
		if model == nil {
			continue
		}
		model.Tags = append(model.Tags, row.Slug)
	}
	return nil
}
