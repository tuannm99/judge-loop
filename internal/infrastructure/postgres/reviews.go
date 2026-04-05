package postgres

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	outport "github.com/tuannm99/judge-loop/internal/port/out"
)

// ReviewRepositoryImpl handles review schedule queries.
type ReviewRepositoryImpl struct{ db *DB }

var _ outport.ReviewRepository = (*ReviewRepositoryImpl)(nil)

// NewReviewRepositoryImpl creates a new ReviewRepositoryImpl.
func NewReviewRepositoryImpl(db *DB) *ReviewRepositoryImpl { return &ReviewRepositoryImpl{db: db} }

// GetDue returns problems due for review today or overdue.
func (s *ReviewRepositoryImpl) GetDue(ctx context.Context, userID uuid.UUID) ([]outport.DueReview, error) {
	var out []outport.DueReview
	err := s.db.Gorm.WithContext(ctx).Raw(`
		SELECT
			p.id AS problem_id,
			p.slug,
			p.title,
			(SELECT MAX(submitted_at)
			 FROM submissions
			 WHERE user_id = $1 AND problem_id = p.id AND status = 'accepted') AS last_solved,
			GREATEST(0, (CURRENT_DATE - r.next_review_at::date)::int) AS days_overdue
		FROM review_schedules r
		JOIN problems p ON p.id = r.problem_id
		WHERE r.user_id = $1
		  AND r.next_review_at <= NOW()
		ORDER BY r.next_review_at ASC`,
		userID,
	).Scan(&out).Error
	if err != nil {
		return nil, fmt.Errorf("get due reviews: %w", err)
	}
	return out, nil
}
