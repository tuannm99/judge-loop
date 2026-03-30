package postgres

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
)

// DueReview is a problem due for spaced repetition review.
type DueReview struct {
	ProblemID   uuid.UUID
	Slug        string
	Title       string
	LastSolved  *time.Time
	DaysOverdue int
}

// ReviewStore handles review schedule queries.
type ReviewStore struct{ db *DB }

// NewReviewStore creates a new ReviewStore.
func NewReviewStore(db *DB) *ReviewStore { return &ReviewStore{db: db} }

// GetDue returns problems due for review today or overdue.
func (s *ReviewStore) GetDue(ctx context.Context, userID uuid.UUID) ([]DueReview, error) {
	var out []DueReview
	err := s.db.Gorm.WithContext(ctx).Raw(`
		SELECT
			p.id,
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
