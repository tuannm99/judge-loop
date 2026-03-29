package storage

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/tuannm99/judge-loop/internal/domain"
)

// problemCols is the SELECT column list for problem rows.
// Enum columns are cast to text to avoid pgx type registration.
const problemCols = `
	id, slug, title, difficulty::text, tags, pattern_tags, provider::text,
	external_id, source_url, estimated_time, created_at, updated_at`

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

// rowScanner is satisfied by both pgx.Row and pgx.Rows.
type rowScanner interface {
	Scan(dest ...any) error
}

func scanProblem(r rowScanner) (domain.Problem, error) {
	var p domain.Problem
	var diff, prov string
	err := r.Scan(
		&p.ID, &p.Slug, &p.Title, &diff,
		&p.Tags, &p.PatternTags, &prov,
		&p.ExternalID, &p.SourceURL, &p.EstimatedTime,
		&p.CreatedAt, &p.UpdatedAt,
	)
	if err != nil {
		return domain.Problem{}, err
	}
	p.Difficulty = domain.Difficulty(diff)
	p.Provider = domain.Provider(prov)
	return p, nil
}

// List returns problems matching the filter with a total count.
func (s *ProblemStore) List(ctx context.Context, f ProblemFilter) ([]domain.Problem, int, error) {
	if f.Limit <= 0 {
		f.Limit = 20
	}

	// convert typed filters to nullable strings for SQL
	var diff, prov *string
	if f.Difficulty != nil {
		d := string(*f.Difficulty)
		diff = &d
	}
	if f.Provider != nil {
		p := string(*f.Provider)
		prov = &p
	}

	const q = `
		SELECT` + problemCols + `
		FROM problems
		WHERE ($1::text IS NULL OR difficulty::text = $1)
		  AND ($2 = ''        OR $2 = ANY(tags))
		  AND ($3 = ''        OR $3 = ANY(pattern_tags))
		  AND ($4::text IS NULL OR provider::text = $4)
		ORDER BY created_at DESC
		LIMIT $5 OFFSET $6`

	rows, err := s.db.Pool.Query(ctx, q, diff, f.Tag, f.Pattern, prov, f.Limit, f.Offset)
	if err != nil {
		return nil, 0, fmt.Errorf("list problems: %w", err)
	}
	defer rows.Close()

	var out []domain.Problem
	for rows.Next() {
		p, err := scanProblem(rows)
		if err != nil {
			return nil, 0, fmt.Errorf("scan problem: %w", err)
		}
		out = append(out, p)
	}
	if err := rows.Err(); err != nil {
		return nil, 0, fmt.Errorf("rows error: %w", err)
	}

	const countQ = `
		SELECT COUNT(*) FROM problems
		WHERE ($1::text IS NULL OR difficulty::text = $1)
		  AND ($2 = ''        OR $2 = ANY(tags))
		  AND ($3 = ''        OR $3 = ANY(pattern_tags))
		  AND ($4::text IS NULL OR provider::text = $4)`

	var total int
	if err := s.db.Pool.QueryRow(ctx, countQ, diff, f.Tag, f.Pattern, prov).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("count problems: %w", err)
	}

	return out, total, nil
}

// GetByID returns a problem by UUID, or nil if not found.
func (s *ProblemStore) GetByID(ctx context.Context, id uuid.UUID) (*domain.Problem, error) {
	row := s.db.Pool.QueryRow(ctx, `SELECT`+problemCols+`FROM problems WHERE id = $1`, id)
	p, err := scanProblem(row)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("get problem by id: %w", err)
	}
	return &p, nil
}

// GetBySlug returns a problem by slug, or nil if not found.
func (s *ProblemStore) GetBySlug(ctx context.Context, slug string) (*domain.Problem, error) {
	row := s.db.Pool.QueryRow(ctx, `SELECT`+problemCols+`FROM problems WHERE slug = $1`, slug)
	p, err := scanProblem(row)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("get problem by slug: %w", err)
	}
	return &p, nil
}

// Suggest returns a random unsolved problem for the user.
// If patterns is non-empty, problems matching those patterns are prioritized.
func (s *ProblemStore) Suggest(ctx context.Context, userID uuid.UUID, patterns []string) (*domain.Problem, error) {
	if patterns == nil {
		patterns = []string{}
	}
	const q = `
		SELECT` + problemCols + `
		FROM problems p
		WHERE p.id NOT IN (
			SELECT DISTINCT problem_id FROM submissions
			WHERE user_id = $1 AND status = 'accepted'
		)
		ORDER BY
			CASE WHEN $2::text[] && p.pattern_tags THEN 0 ELSE 1 END,
			RANDOM()
		LIMIT 1`

	row := s.db.Pool.QueryRow(ctx, q, userID, patterns)
	p, err := scanProblem(row)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("suggest problem: %w", err)
	}
	return &p, nil
}
