package postgres

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/tuannm99/judge-loop/internal/domain"
	outport "github.com/tuannm99/judge-loop/internal/port/out"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// SessionRepositoryImpl handles DailySession and TimerSession queries.
type SessionRepositoryImpl struct{ db *DB }

var _ outport.SessionRepository = (*SessionRepositoryImpl)(nil)

// NewSessionRepositoryImpl creates a new SessionRepositoryImpl.
func NewSessionRepositoryImpl(db *DB) *SessionRepositoryImpl { return &SessionRepositoryImpl{db: db} }

// GetOrCreateToday returns today's DailySession for the user, creating it if needed.
func (s *SessionRepositoryImpl) GetOrCreateToday(ctx context.Context, userID uuid.UUID) (*domain.DailySession, error) {
	today := time.Now().UTC().Truncate(24 * time.Hour)
	model := dailySessionModel{
		ID:     uuid.New(),
		UserID: userID,
		Date:   today,
	}
	if err := s.db.Gorm.WithContext(ctx).
		Clauses(clause.OnConflict{Columns: []clause.Column{{Name: "user_id"}, {Name: "date"}}, DoNothing: true}).
		Create(&model).Error; err != nil {
		return nil, fmt.Errorf("create daily session: %w", err)
	}

	if err := s.db.Gorm.WithContext(ctx).Where("user_id = ? AND date = ?", userID, today).Take(&model).Error; err != nil {
		return nil, fmt.Errorf("get or create daily session: %w", err)
	}
	ds := model.toDomain()
	return &ds, nil
}

// RecordSubmission updates today's DailySession after a submission is evaluated.
// It always increments attempted_count; it increments solved_count only when accepted=true.
func (s *SessionRepositoryImpl) RecordSubmission(ctx context.Context, userID uuid.UUID, accepted bool) error {
	solved := 0
	if accepted {
		solved = 1
	}
	today := time.Now().UTC().Truncate(24 * time.Hour)

	return s.db.Gorm.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		model := dailySessionModel{
			ID:             uuid.New(),
			UserID:         userID,
			Date:           today,
			AttemptedCount: 1,
			SolvedCount:    solved,
		}
		if err := tx.Clauses(clause.OnConflict{Columns: []clause.Column{{Name: "user_id"}, {Name: "date"}}, DoNothing: true}).Create(&model).Error; err != nil {
			return fmt.Errorf("insert daily session: %w", err)
		}
		if err := tx.Model(&dailySessionModel{}).
			Where("user_id = ? AND date = ?", userID, today).
			Updates(map[string]any{
				"attempted_count": gorm.Expr("attempted_count + ?", 1),
				"solved_count":    gorm.Expr("solved_count + ?", solved),
				"updated_at":      gorm.Expr("NOW()"),
			}).Error; err != nil {
			return fmt.Errorf("record submission update: %w", err)
		}
		return nil
	})
}

// StartTimer stops any existing active timer and starts a new one.
func (s *SessionRepositoryImpl) StartTimer(
	ctx context.Context,
	userID uuid.UUID,
	problemID *uuid.UUID,
) (*domain.TimerSession, error) {
	now := time.Now()
	if err := s.db.Gorm.WithContext(ctx).
		Model(&timerSessionModel{}).
		Where("user_id = ? AND ended_at IS NULL", userID).
		Updates(map[string]any{
			"ended_at":     now,
			"elapsed_secs": gorm.Expr("GREATEST(0, EXTRACT(EPOCH FROM (? - started_at))::int)", now),
		}).Error; err != nil {
		return nil, fmt.Errorf("stop prior timer: %w", err)
	}

	model := timerSessionModel{
		ID:        uuid.New(),
		UserID:    userID,
		ProblemID: problemID,
		StartedAt: now,
	}
	if err := s.db.Gorm.WithContext(ctx).Create(&model).Error; err != nil {
		return nil, fmt.Errorf("start timer: %w", err)
	}
	ts := model.toDomain()
	return &ts, nil
}

// StopTimer stops the currently active timer for the user and records elapsed
// time in today's DailySession.
func (s *SessionRepositoryImpl) StopTimer(ctx context.Context, userID uuid.UUID) (*domain.TimerSession, error) {
	var model timerSessionModel
	err := s.db.Gorm.WithContext(ctx).
		Where("user_id = ? AND ended_at IS NULL", userID).
		Order("started_at DESC").
		Take(&model).Error
	if err == gorm.ErrRecordNotFound {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("stop timer lookup: %w", err)
	}

	now := time.Now()
	elapsed := int(now.Sub(model.StartedAt).Seconds())
	if elapsed < 0 {
		elapsed = 0
	}

	model.EndedAt = &now
	model.ElapsedSecs = elapsed
	if err := s.db.Gorm.WithContext(ctx).Save(&model).Error; err != nil {
		return nil, fmt.Errorf("stop timer save: %w", err)
	}

	today := now.UTC().Truncate(24 * time.Hour)
	if err := s.db.Gorm.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		session := dailySessionModel{
			ID:     uuid.New(),
			UserID: userID,
			Date:   today,
		}
		if err := tx.Clauses(clause.OnConflict{Columns: []clause.Column{{Name: "user_id"}, {Name: "date"}}, DoNothing: true}).Create(&session).Error; err != nil {
			return err
		}
		return tx.Model(&dailySessionModel{}).
			Where("user_id = ? AND date = ?", userID, today).
			Updates(map[string]any{
				"time_spent_secs": gorm.Expr("time_spent_secs + ?", elapsed),
				"updated_at":      gorm.Expr("NOW()"),
			}).Error
	}); err != nil {
		return nil, fmt.Errorf("stop timer update daily session: %w", err)
	}

	ts := model.toDomain()
	return &ts, nil
}

// ActiveTimer returns the currently running timer for the user, or nil.
func (s *SessionRepositoryImpl) ActiveTimer(ctx context.Context, userID uuid.UUID) (*domain.TimerSession, error) {
	var model timerSessionModel
	err := s.db.Gorm.WithContext(ctx).
		Where("user_id = ? AND ended_at IS NULL", userID).
		Order("started_at DESC").
		Take(&model).Error
	if err == gorm.ErrRecordNotFound {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("active timer: %w", err)
	}
	ts := model.toDomain()
	return &ts, nil
}

// ElapsedNow returns seconds elapsed for an active (not-yet-stopped) timer.
func ElapsedNow(ts *domain.TimerSession) int {
	if ts == nil {
		return 0
	}
	if ts.EndedAt != nil {
		return ts.ElapsedSecs
	}
	return int(time.Since(ts.StartedAt).Seconds())
}
