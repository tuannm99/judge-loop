package postgres

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/tuannm99/judge-loop/internal/domain"
	outport "github.com/tuannm99/judge-loop/internal/port/out"
	"gorm.io/gorm"
)

// SubmissionStore handles all submission queries.
type SubmissionStore struct{ db *DB }

var _ outport.SubmissionRepository = (*SubmissionStore)(nil)

// NewSubmissionStore creates a new SubmissionStore.
func NewSubmissionStore(db *DB) *SubmissionStore { return &SubmissionStore{db: db} }

// Create inserts a new submission with status=pending and returns the ID.
func (s *SubmissionStore) Create(ctx context.Context, sub *domain.Submission) error {
	sub.ID = uuid.New()
	sub.Status = domain.StatusPending
	sub.SubmittedAt = time.Now()

	model := submissionFromDomain(*sub)
	if err := s.db.Gorm.WithContext(ctx).Create(&model).Error; err != nil {
		return fmt.Errorf("create submission: %w", err)
	}
	return nil
}

// GetByID returns a submission by UUID, or nil if not found.
func (s *SubmissionStore) GetByID(ctx context.Context, id uuid.UUID) (*domain.Submission, error) {
	var model submissionModel
	err := s.db.Gorm.WithContext(ctx).First(&model, "id = ?", id).Error
	if err == gorm.ErrRecordNotFound {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("get submission: %w", err)
	}
	sub := model.toDomain()
	return &sub, nil
}

// UpdateVerdict writes the final verdict back to the submission row.
func (s *SubmissionStore) UpdateVerdict(
	ctx context.Context,
	id uuid.UUID,
	status, verdict string,
	passed, total int,
	runtimeMS int64,
	errMsg string,
	evaluatedAt *time.Time,
) error {
	err := s.db.Gorm.WithContext(ctx).
		Model(&submissionModel{}).
		Where("id = ?", id).
		Updates(map[string]any{
			"status":        status,
			"verdict":       verdict,
			"passed_cases":  passed,
			"total_cases":   total,
			"runtime_ms":    runtimeMS,
			"error_message": errMsg,
			"evaluated_at":  evaluatedAt,
		}).Error
	if err != nil {
		return fmt.Errorf("update verdict: %w", err)
	}
	return nil
}

// ListByUser returns the user's submission history, ordered by most recent.
func (s *SubmissionStore) ListByUser(
	ctx context.Context,
	userID uuid.UUID,
	problemID *uuid.UUID,
	limit, offset int,
) ([]domain.Submission, error) {
	if limit <= 0 {
		limit = 20
	}

	q := s.db.Gorm.WithContext(ctx).
		Model(&submissionModel{}).
		Where("user_id = ?", userID).
		Order("submitted_at DESC").
		Limit(limit).
		Offset(offset)
	if problemID != nil {
		q = q.Where("problem_id = ?", *problemID)
	}

	var models []submissionModel
	if err := q.Find(&models).Error; err != nil {
		return nil, fmt.Errorf("list submissions: %w", err)
	}

	out := make([]domain.Submission, 0, len(models))
	for _, model := range models {
		out = append(out, model.toDomain())
	}
	return out, nil
}
