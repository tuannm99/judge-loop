package postgres

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	outport "github.com/tuannm99/judge-loop/internal/port/out"
)

// GetStreak returns the user's current and all-time longest streak.
// A streak day is any day with at least one accepted submission.
func (s *SessionRepositoryImpl) GetStreak(ctx context.Context, userID uuid.UUID) (outport.StreakInfo, error) {
	var rows []dailySessionModel
	if err := s.db.Gorm.WithContext(ctx).
		Model(&dailySessionModel{}).
		Select("date").
		Where("user_id = ? AND solved_count > 0", userID).
		Order("date DESC").
		Find(&rows).Error; err != nil {
		return outport.StreakInfo{}, fmt.Errorf("streak query: %w", err)
	}

	dates := make([]time.Time, 0, len(rows))
	for _, row := range rows {
		dates = append(dates, row.Date.UTC().Truncate(24*time.Hour))
	}

	return computeStreak(dates), nil
}

// computeStreak calculates current and longest streak from a DESC-sorted list of dates.
func computeStreak(dates []time.Time) outport.StreakInfo {
	if len(dates) == 0 {
		return outport.StreakInfo{}
	}

	info := outport.StreakInfo{LastPracticed: &dates[0]}

	today := time.Now().UTC().Truncate(24 * time.Hour)
	yesterday := today.AddDate(0, 0, -1)

	// current streak: count consecutive days from today or yesterday backwards
	if dates[0].Equal(today) || dates[0].Equal(yesterday) {
		info.Current = 1
		for i := 1; i < len(dates); i++ {
			expected := dates[i-1].AddDate(0, 0, -1)
			if dates[i].Equal(expected) {
				info.Current++
			} else {
				break
			}
		}
	}

	// longest streak: scan all dates for max consecutive run
	run := 1
	longest := 1
	for i := 1; i < len(dates); i++ {
		expected := dates[i-1].AddDate(0, 0, -1)
		if dates[i].Equal(expected) {
			run++
			if run > longest {
				longest = run
			}
		} else {
			run = 1
		}
	}
	info.Longest = longest
	if info.Current > info.Longest {
		info.Longest = info.Current
	}

	return info
}
