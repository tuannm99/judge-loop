package postgres

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/lib/pq"
	"github.com/tuannm99/judge-loop/internal/domain"
	"gorm.io/gorm"
)

// ProblemStore handles all problem queries.
type ProblemStore struct{ db *DB }

// NewProblemStore creates a new ProblemStore.
func NewProblemStore(db *DB) *ProblemStore { return &ProblemStore{db: db} }

// ProblemFilter holds optional filters for listing problems.
// Nil pointer fields mean "no filter".
type ProblemFilter struct {
	Difficulty *domain.Difficulty
	Tag        string
	Pattern    string
	Provider   *domain.Provider
	Limit      int
	Offset     int
}

// List returns problems matching the filter with a total count.
func (s *ProblemStore) List(ctx context.Context, f ProblemFilter) ([]domain.Problem, int, error) {
	if f.Limit <= 0 {
		f.Limit = 20
	}

	q := s.db.Gorm.WithContext(ctx).Model(&problemModel{})
	if f.Difficulty != nil {
		q = q.Where("difficulty = ?", string(*f.Difficulty))
	}
	if f.Provider != nil {
		q = q.Where("provider = ?", string(*f.Provider))
	}
	if f.Tag != "" {
		q = q.Where("? = ANY(tags)", f.Tag)
	}
	if f.Pattern != "" {
		q = q.Where("? = ANY(pattern_tags)", f.Pattern)
	}

	var total int64
	if err := q.Count(&total).Error; err != nil {
		return nil, 0, fmt.Errorf("count problems: %w", err)
	}

	var models []problemModel
	if err := q.Order("created_at DESC").Limit(f.Limit).Offset(f.Offset).Find(&models).Error; err != nil {
		return nil, 0, fmt.Errorf("list problems: %w", err)
	}

	out := make([]domain.Problem, 0, len(models))
	for _, model := range models {
		out = append(out, model.toDomain())
	}
	return out, int(total), nil
}

// GetByID returns a problem by UUID, or nil if not found.
func (s *ProblemStore) GetByID(ctx context.Context, id uuid.UUID) (*domain.Problem, error) {
	var model problemModel
	err := s.db.Gorm.WithContext(ctx).First(&model, "id = ?", id).Error
	if err == gorm.ErrRecordNotFound {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("get problem by id: %w", err)
	}
	p := model.toDomain()
	return &p, nil
}

// GetBySlug returns a problem by slug, or nil if not found.
func (s *ProblemStore) GetBySlug(ctx context.Context, slug string) (*domain.Problem, error) {
	var model problemModel
	err := s.db.Gorm.WithContext(ctx).First(&model, "slug = ?", slug).Error
	if err == gorm.ErrRecordNotFound {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("get problem by slug: %w", err)
	}
	p := model.toDomain()
	return &p, nil
}

// Suggest returns a random unsolved problem for the user.
// If patterns is non-empty, problems matching those patterns are prioritized.
func (s *ProblemStore) Suggest(ctx context.Context, userID uuid.UUID, patterns []string) (*domain.Problem, error) {
	if patterns == nil {
		patterns = []string{}
	}

	q := s.db.Gorm.WithContext(ctx).
		Model(&problemModel{}).
		Where("id NOT IN (?)",
			s.db.Gorm.WithContext(ctx).
				Model(&submissionModel{}).
				Select("DISTINCT problem_id").
				Where("user_id = ? AND status = ?", userID, string(domain.StatusAccepted)),
		)

	if len(patterns) > 0 {
		q = q.Order(gorm.Expr("CASE WHEN ?::text[] && pattern_tags THEN 0 ELSE 1 END", pq.Array(patterns)))
	}

	var model problemModel
	err := q.Order("RANDOM()").Limit(1).Take(&model).Error
	if err == gorm.ErrRecordNotFound {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("suggest problem: %w", err)
	}
	p := model.toDomain()
	return &p, nil
}
