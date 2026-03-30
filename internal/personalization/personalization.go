// Package personalization generates daily missions and computes weak patterns.
package personalization

import (
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/tuannm99/judge-loop/internal/domain"
	"github.com/tuannm99/judge-loop/internal/storage"
)

const (
	defaultDailyGoal    = 2
	weakPatternThreshold = 0.5 // score below this is "weak"
	minAttemptsForWeak  = 2    // need at least this many attempts before flagging weak
	optionalTaskCount   = 2
)

// Input gathers all data needed to generate a daily mission.
type Input struct {
	PatternScores map[string]float64 // pattern → accepted/attempted ratio
	Unsolved      []domain.Problem   // problems not yet accepted by the user
	DueReviews    []storage.DueReview
	DailyGoal     int // 0 = use default
}

// Generate builds a DailyMission from the given input.
// It selects required tasks (weak-pattern problems first), review tasks, and optional tasks.
func Generate(userID uuid.UUID, in Input) domain.DailyMission {
	goal := in.DailyGoal
	if goal <= 0 {
		goal = defaultDailyGoal
	}

	mission := domain.DailyMission{
		ID:          uuid.New(),
		UserID:      userID,
		Date:        time.Now().UTC().Truncate(24 * time.Hour),
		GeneratedAt: time.Now(),
	}

	// Review tasks from spaced repetition schedule.
	for _, r := range in.DueReviews {
		reason := "due for review"
		if r.DaysOverdue > 0 {
			reason = fmt.Sprintf("overdue by %d day(s)", r.DaysOverdue)
		}
		mission.ReviewTasks = append(mission.ReviewTasks, domain.MissionTask{
			ProblemID: r.ProblemID,
			Slug:      r.Slug,
			Title:     r.Title,
			Reason:    reason,
			Priority:  10 + r.DaysOverdue,
		})
	}

	// Identify weak patterns from score map.
	weak := weakPatterns(in.PatternScores)

	// Required tasks: weak-pattern problems first, then fill with any unsolved.
	added := make(map[uuid.UUID]bool)
	for _, p := range in.Unsolved {
		if len(mission.RequiredTasks) >= goal {
			break
		}
		if match := firstMatch(p.PatternTags, weak); match != "" {
			mission.RequiredTasks = append(mission.RequiredTasks, domain.MissionTask{
				ProblemID:  p.ID,
				Slug:       p.Slug,
				Title:      p.Title,
				Difficulty: p.Difficulty,
				Reason:     "weak pattern: " + match,
				Priority:   5,
			})
			added[p.ID] = true
		}
	}
	for _, p := range in.Unsolved {
		if len(mission.RequiredTasks) >= goal {
			break
		}
		if added[p.ID] {
			continue
		}
		mission.RequiredTasks = append(mission.RequiredTasks, domain.MissionTask{
			ProblemID:  p.ID,
			Slug:       p.Slug,
			Title:      p.Title,
			Difficulty: p.Difficulty,
			Reason:     "unsolved",
			Priority:   3,
		})
		added[p.ID] = true
	}

	// Optional tasks: next unsolved problems beyond the required set.
	for _, p := range in.Unsolved {
		if len(mission.OptionalTasks) >= optionalTaskCount {
			break
		}
		if added[p.ID] {
			continue
		}
		mission.OptionalTasks = append(mission.OptionalTasks, domain.MissionTask{
			ProblemID:  p.ID,
			Slug:       p.Slug,
			Title:      p.Title,
			Difficulty: p.Difficulty,
			Reason:     "optional challenge",
			Priority:   1,
		})
		added[p.ID] = true
	}

	return mission
}

// WeakPatterns returns pattern tags with score below the threshold from pattern scores.
func WeakPatterns(scores map[string]float64) []string {
	return weakPatterns(scores)
}

func weakPatterns(scores map[string]float64) []string {
	var out []string
	for pattern, score := range scores {
		if score < weakPatternThreshold {
			out = append(out, pattern)
		}
	}
	return out
}

func firstMatch(tags, targets []string) string {
	set := make(map[string]bool, len(targets))
	for _, t := range targets {
		set[t] = true
	}
	for _, tag := range tags {
		if set[tag] {
			return tag
		}
	}
	return ""
}
