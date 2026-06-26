package postgres

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/google/uuid"
	outport "github.com/tuannm99/judge-loop/internal/port/out"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

const (
	evaluationJobStatusPending = "pending"
	evaluationJobStatusRunning = "running"
	evaluationJobStatusDone    = "done"
	evaluationJobStatusFailed  = "failed"

	evaluationJobVisibilityTimeout = 5 * time.Minute
)

// EvaluationJobRepositoryImpl stores submission evaluation jobs in Postgres.
type EvaluationJobRepositoryImpl struct{ db *DB }

var (
	_ outport.EvaluationPublisher = (*EvaluationJobRepositoryImpl)(nil)
	_ outport.EvaluationJobQueue  = (*EvaluationJobRepositoryImpl)(nil)
)

func NewEvaluationJobRepositoryImpl(db *DB) *EvaluationJobRepositoryImpl {
	return &EvaluationJobRepositoryImpl{db: db}
}

func (s *EvaluationJobRepositoryImpl) PublishEvaluation(job outport.EvaluateSubmissionJob) error {
	submissionID, err := uuid.Parse(job.SubmissionID)
	if err != nil {
		return fmt.Errorf("parse submission_id: %w", err)
	}
	userID, err := uuid.Parse(job.UserID)
	if err != nil {
		return fmt.Errorf("parse user_id: %w", err)
	}

	model := evaluationJobModel{
		SubmissionID: submissionID,
		UserID:       userID,
		Status:       evaluationJobStatusPending,
		MaxAttempts:  3,
		AvailableAt:  time.Now().UTC(),
	}
	err = s.db.Gorm.Clauses(clause.OnConflict{
		Columns: []clause.Column{{Name: "submission_id"}},
		DoUpdates: clause.Assignments(map[string]any{
			"user_id":      model.UserID,
			"status":       evaluationJobStatusPending,
			"attempts":     0,
			"available_at": gorm.Expr("NOW()"),
			"locked_at":    nil,
			"locked_by":    "",
			"last_error":   "",
			"updated_at":   gorm.Expr("NOW()"),
		}),
	}).Create(&model).Error
	if err != nil {
		log.Printf("enqueue evaluate job for %s: %v", job.SubmissionID, err)
		return fmt.Errorf("enqueue evaluate job: %w", err)
	}
	return nil
}

func (s *EvaluationJobRepositoryImpl) ClaimEvaluationJob(
	ctx context.Context,
	workerID string,
) (*outport.EvaluationJob, error) {
	var model evaluationJobModel
	err := s.db.Gorm.WithContext(ctx).Raw(
		`
		UPDATE evaluation_jobs
		SET status = ?,
		    attempts = attempts + 1,
		    locked_at = NOW(),
		    locked_by = ?,
		    updated_at = NOW()
		WHERE id = (
			SELECT id
			FROM evaluation_jobs
			WHERE (
				status = ?
				AND available_at <= NOW()
			) OR (
				status = ?
				AND locked_at < NOW() - make_interval(secs => ?)
				AND attempts < max_attempts
			)
			ORDER BY created_at ASC
			FOR UPDATE SKIP LOCKED
			LIMIT 1
		)
		RETURNING *
	`,
		evaluationJobStatusRunning,
		workerID,
		evaluationJobStatusPending,
		evaluationJobStatusRunning,
		int(evaluationJobVisibilityTimeout.Seconds()),
	).Scan(&model).Error
	if err != nil {
		return nil, fmt.Errorf("claim evaluation job: %w", err)
	}
	if model.ID == uuid.Nil {
		return nil, nil
	}
	job := model.toPort()
	return &job, nil
}

func (s *EvaluationJobRepositoryImpl) CompleteEvaluationJob(ctx context.Context, id uuid.UUID) error {
	err := s.db.Gorm.WithContext(ctx).
		Model(&evaluationJobModel{}).
		Where("id = ?", id).
		Updates(map[string]any{
			"status":       evaluationJobStatusDone,
			"locked_at":    nil,
			"locked_by":    "",
			"last_error":   "",
			"updated_at":   gorm.Expr("NOW()"),
			"available_at": gorm.Expr("NOW()"),
		}).Error
	if err != nil {
		return fmt.Errorf("complete evaluation job %s: %w", id, err)
	}
	return nil
}

func (s *EvaluationJobRepositoryImpl) FailEvaluationJob(ctx context.Context, id uuid.UUID, errMsg string) error {
	var model evaluationJobModel
	err := s.db.Gorm.WithContext(ctx).First(&model, "id = ?", id).Error
	if err == gorm.ErrRecordNotFound {
		return nil
	}
	if err != nil {
		return fmt.Errorf("load failed evaluation job %s: %w", id, err)
	}

	nextStatus := evaluationJobStatusPending
	if model.Attempts >= model.MaxAttempts {
		nextStatus = evaluationJobStatusFailed
	}
	backoff := time.Duration(model.Attempts*model.Attempts) * time.Second
	if backoff < time.Second {
		backoff = time.Second
	}
	if backoff > time.Minute {
		backoff = time.Minute
	}

	err = s.db.Gorm.WithContext(ctx).
		Model(&evaluationJobModel{}).
		Where("id = ?", id).
		Updates(map[string]any{
			"status":       nextStatus,
			"available_at": time.Now().UTC().Add(backoff),
			"locked_at":    nil,
			"locked_by":    "",
			"last_error":   errMsg,
			"updated_at":   gorm.Expr("NOW()"),
		}).Error
	if err != nil {
		return fmt.Errorf("fail evaluation job %s: %w", id, err)
	}
	return nil
}
