package postgres

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/google/uuid"
	"github.com/tuannm99/judge-loop/internal/domain"
	outport "github.com/tuannm99/judge-loop/internal/port/out"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// MissionRepositoryImpl handles daily mission persistence.
type MissionRepositoryImpl struct{ db *DB }

var _ outport.MissionRepository = (*MissionRepositoryImpl)(nil)

// NewMissionRepositoryImpl creates a new MissionRepositoryImpl.
func NewMissionRepositoryImpl(db *DB) *MissionRepositoryImpl { return &MissionRepositoryImpl{db: db} }

// GetToday returns today's cached mission for the user, or nil if none exists yet.
func (s *MissionRepositoryImpl) GetToday(ctx context.Context, userID uuid.UUID) (*domain.DailyMission, error) {
	var model dailyMissionModel
	err := s.db.Gorm.WithContext(ctx).
		Where("user_id = ? AND date = CURRENT_DATE", userID).
		Take(&model).Error
	if err == gorm.ErrRecordNotFound {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("get mission: %w", err)
	}

	var m domain.DailyMission
	m.ID = model.ID
	m.UserID = model.UserID
	m.Date = model.Date
	m.GeneratedAt = model.GeneratedAt
	if err := json.Unmarshal(model.RequiredTasks, &m.RequiredTasks); err != nil {
		log.Printf("unmarshal required_tasks for mission %s: %v", model.ID, err)
	}
	if err := json.Unmarshal(model.OptionalTasks, &m.OptionalTasks); err != nil {
		log.Printf("unmarshal optional_tasks for mission %s: %v", model.ID, err)
	}
	if err := json.Unmarshal(model.ReviewTasks, &m.ReviewTasks); err != nil {
		log.Printf("unmarshal review_tasks for mission %s: %v", model.ID, err)
	}
	return &m, nil
}

// Save inserts or replaces today's mission for the user.
func (s *MissionRepositoryImpl) Save(ctx context.Context, m domain.DailyMission) error {
	model, err := missionModelFromDomain(m)
	if err != nil {
		return fmt.Errorf("build mission model: %w", err)
	}

	if err := s.db.Gorm.WithContext(ctx).Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "user_id"}, {Name: "date"}},
		DoUpdates: clause.AssignmentColumns([]string{"required_tasks", "optional_tasks", "review_tasks", "generated_at"}),
	}).Create(&model).Error; err != nil {
		return fmt.Errorf("save mission: %w", err)
	}
	return nil
}

// PerformanceRepositoryImpl computes user performance metrics from submissions.
type PerformanceRepositoryImpl struct{ db *DB }

var _ outport.PerformanceRepository = (*PerformanceRepositoryImpl)(nil)

// NewPerformanceRepositoryImpl creates a new PerformanceRepositoryImpl.
func NewPerformanceRepositoryImpl(db *DB) *PerformanceRepositoryImpl {
	return &PerformanceRepositoryImpl{db: db}
}

// PatternScoreRow holds per-pattern accepted/attempted counts.
type PatternScoreRow struct {
	Pattern   string
	Accepted  int
	Attempted int
}

// GetPatternScores returns a map of pattern tag -> score (accepted/attempted).
func (s *PerformanceRepositoryImpl) GetPatternScores(ctx context.Context, userID uuid.UUID) (map[string]float64, error) {
	var rows []PatternScoreRow
	err := s.db.Gorm.WithContext(ctx).Raw(`
		SELECT
			pl.slug                                       AS pattern,
			COUNT(*) FILTER (WHERE s.status = 'accepted') AS accepted,
			COUNT(*)                                      AS attempted
		FROM submissions s
		JOIN problems p ON p.id = s.problem_id
		JOIN problem_label_links pll ON pll.problem_id = p.id
		JOIN problem_labels pl ON pl.id = pll.problem_label_id
		WHERE s.user_id = $1
		  AND pl.kind = 'pattern'
		GROUP BY pl.slug
		ORDER BY pattern`,
		userID,
	).Scan(&rows).Error
	if err != nil {
		return nil, fmt.Errorf("pattern scores: %w", err)
	}

	scores := make(map[string]float64, len(rows))
	for _, r := range rows {
		if r.Attempted > 0 {
			scores[r.Pattern] = float64(r.Accepted) / float64(r.Attempted)
		}
	}
	return scores, nil
}

// PerformanceStats holds aggregate submission statistics for a user.
type PerformanceStats struct {
	TotalAttempts int     `gorm:"column:total_attempts"`
	AcceptedCount int     `gorm:"column:accepted_count"`
	AvgSolveTime  float64 `gorm:"column:avg_solve_minutes"`
}

// GetStats returns aggregate performance statistics for the user.
func (s *PerformanceRepositoryImpl) GetStats(ctx context.Context, userID uuid.UUID) (PerformanceStats, error) {
	var st PerformanceStats
	err := s.db.Gorm.WithContext(ctx).Raw(`
		SELECT
			COUNT(*)                                       AS total_attempts,
			COUNT(*) FILTER (WHERE s.status = 'accepted') AS accepted_count,
			COALESCE(AVG(CASE WHEN t.elapsed_secs > 0 THEN t.elapsed_secs / 60.0 END), 0) AS avg_solve_minutes
		FROM submissions s
		LEFT JOIN timer_sessions t ON t.id = s.session_id
		WHERE s.user_id = $1`,
		userID,
	).Scan(&st).Error
	if err != nil {
		return PerformanceStats{}, fmt.Errorf("get stats: %w", err)
	}
	return st, nil
}

// GetUnsolved returns problems the user has not yet accepted, in random order.
func (s *ProblemRepositoryImpl) GetUnsolved(ctx context.Context, userID uuid.UUID, limit int) ([]domain.Problem, error) {
	if limit <= 0 {
		limit = 10
	}

	var models []problemModel
	err := s.db.Gorm.WithContext(ctx).
		Model(&problemModel{}).
		Where("problems.id NOT IN (?)",
			s.db.Gorm.WithContext(ctx).
				Model(&submissionModel{}).
				Select("DISTINCT problem_id").
				Where("user_id = ? AND status = ?", userID, string(domain.StatusAccepted)),
		).
		Order("RANDOM()").
		Limit(limit).
		Find(&models).Error
	if err != nil {
		return nil, fmt.Errorf("get unsolved: %w", err)
	}
	modelPtrs := make([]*problemModel, 0, len(models))
	for i := range models {
		modelPtrs = append(modelPtrs, &models[i])
	}
	if err := s.loadProblemLabels(ctx, modelPtrs); err != nil {
		return nil, err
	}

	out := make([]domain.Problem, 0, len(models))
	for _, model := range models {
		out = append(out, model.toDomain())
	}
	return out, nil
}

// UpsertReview creates or advances the review schedule for a problem.
func (s *ReviewRepositoryImpl) Upsert(ctx context.Context, userID, problemID uuid.UUID) error {
	now := time.Now()

	return s.db.Gorm.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var existing reviewScheduleModel
		err := tx.Where("user_id = ? AND problem_id = ?", userID, problemID).Take(&existing).Error
		if err == gorm.ErrRecordNotFound {
			model := reviewScheduleModel{
				ID:           uuid.New(),
				UserID:       userID,
				ProblemID:    problemID,
				NextReviewAt: now.Add(24 * time.Hour),
				IntervalDays: 1,
				ReviewCount:  1,
			}
			if err := tx.Create(&model).Error; err != nil {
				return fmt.Errorf("create review schedule: %w", err)
			}
			return nil
		}
		if err != nil {
			return fmt.Errorf("load review schedule: %w", err)
		}

		nextInterval := 1
		switch existing.ReviewCount {
		case 1:
			nextInterval = 3
		case 2:
			nextInterval = 7
		default:
			if existing.ReviewCount >= 3 {
				nextInterval = 14
			}
		}

		if err := tx.Model(&reviewScheduleModel{}).
			Where("user_id = ? AND problem_id = ?", userID, problemID).
			Updates(map[string]any{
				"interval_days":  nextInterval,
				"next_review_at": now.Add(time.Duration(nextInterval) * 24 * time.Hour),
				"review_count":   existing.ReviewCount + 1,
				"updated_at":     now,
			}).Error; err != nil {
			return fmt.Errorf("update review schedule: %w", err)
		}
		return nil
	})
}

// LastSyncedAt returns when the registry was last synced, or zero time.
func (s *RegistryRepositoryImpl) LastSyncedAt(ctx context.Context) time.Time {
	row, _ := s.GetLatest(ctx)
	if row == nil {
		return time.Time{}
	}
	return row.SyncedAt
}
