package postgres

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
	"github.com/tuannm99/judge-loop/internal/domain"
)

type problemModel struct {
	ID            uuid.UUID `gorm:"type:uuid;default:gen_random_uuid();primaryKey"`
	Slug          string    `gorm:"column:slug"`
	Title         string    `gorm:"column:title"`
	Difficulty    string    `gorm:"column:difficulty"`
	Provider      string    `gorm:"column:provider"`
	ExternalID    string    `gorm:"column:external_id"`
	SourceURL     string    `gorm:"column:source_url"`
	EstimatedTime int       `gorm:"column:estimated_time"`
	StarterCode   []byte    `gorm:"column:starter_code"`
	CreatedAt     time.Time `gorm:"column:created_at"`
	UpdatedAt     time.Time `gorm:"column:updated_at"`

	Tags        []string `gorm:"-"`
	PatternTags []string `gorm:"-"`
}

func (problemModel) TableName() string { return "problems" }

func (m problemModel) toDomain() domain.Problem {
	starterCode := map[string]string{}
	if len(m.StarterCode) > 0 {
		_ = json.Unmarshal(m.StarterCode, &starterCode)
	}

	return domain.Problem{
		ID:            m.ID,
		Slug:          m.Slug,
		Title:         m.Title,
		Difficulty:    domain.Difficulty(m.Difficulty),
		Tags:          append([]string(nil), m.Tags...),
		PatternTags:   append([]string(nil), m.PatternTags...),
		Provider:      domain.Provider(m.Provider),
		ExternalID:    m.ExternalID,
		SourceURL:     m.SourceURL,
		EstimatedTime: m.EstimatedTime,
		StarterCode:   starterCode,
		CreatedAt:     m.CreatedAt,
		UpdatedAt:     m.UpdatedAt,
	}
}

type problemLabelModel struct {
	ID        uuid.UUID `gorm:"type:uuid;default:gen_random_uuid();primaryKey"`
	Kind      string    `gorm:"column:kind"`
	Slug      string    `gorm:"column:slug"`
	Name      string    `gorm:"column:name"`
	CreatedAt time.Time `gorm:"column:created_at"`
	UpdatedAt time.Time `gorm:"column:updated_at"`
}

func (problemLabelModel) TableName() string { return "problem_labels" }

func (m problemLabelModel) toDomain() domain.ProblemLabel {
	return domain.ProblemLabel{
		ID:        m.ID,
		Kind:      m.Kind,
		Slug:      m.Slug,
		Name:      m.Name,
		CreatedAt: m.CreatedAt,
		UpdatedAt: m.UpdatedAt,
	}
}

type problemLabelLinkModel struct {
	ProblemID      uuid.UUID `gorm:"column:problem_id;primaryKey"`
	ProblemLabelID uuid.UUID `gorm:"column:problem_label_id;primaryKey"`
}

func (problemLabelLinkModel) TableName() string { return "problem_label_links" }

type submissionModel struct {
	ID           uuid.UUID  `gorm:"type:uuid;default:gen_random_uuid();primaryKey"`
	UserID       uuid.UUID  `gorm:"column:user_id"`
	ProblemID    uuid.UUID  `gorm:"column:problem_id"`
	SessionID    *uuid.UUID `gorm:"column:session_id"`
	Language     string     `gorm:"column:language"`
	Code         string     `gorm:"column:code"`
	Status       string     `gorm:"column:status"`
	Verdict      string     `gorm:"column:verdict"`
	PassedCases  int        `gorm:"column:passed_cases"`
	TotalCases   int        `gorm:"column:total_cases"`
	RuntimeMS    int64      `gorm:"column:runtime_ms"`
	ErrorMessage string     `gorm:"column:error_message"`
	SubmittedAt  time.Time  `gorm:"column:submitted_at"`
	EvaluatedAt  *time.Time `gorm:"column:evaluated_at"`
}

func (submissionModel) TableName() string { return "submissions" }

func submissionFromDomain(sub domain.Submission) submissionModel {
	return submissionModel{
		ID:           sub.ID,
		UserID:       sub.UserID,
		ProblemID:    sub.ProblemID,
		SessionID:    sub.SessionID,
		Language:     string(sub.Language),
		Code:         sub.Code,
		Status:       string(sub.Status),
		Verdict:      string(sub.Verdict),
		PassedCases:  sub.PassedCases,
		TotalCases:   sub.TotalCases,
		RuntimeMS:    sub.RuntimeMS,
		ErrorMessage: sub.ErrorMessage,
		SubmittedAt:  sub.SubmittedAt,
		EvaluatedAt:  sub.EvaluatedAt,
	}
}

func (m submissionModel) toDomain() domain.Submission {
	return domain.Submission{
		ID:           m.ID,
		UserID:       m.UserID,
		ProblemID:    m.ProblemID,
		SessionID:    m.SessionID,
		Language:     domain.Language(m.Language),
		Code:         m.Code,
		Status:       domain.SubmissionStatus(m.Status),
		Verdict:      domain.Verdict(m.Verdict),
		PassedCases:  m.PassedCases,
		TotalCases:   m.TotalCases,
		RuntimeMS:    m.RuntimeMS,
		ErrorMessage: m.ErrorMessage,
		SubmittedAt:  m.SubmittedAt,
		EvaluatedAt:  m.EvaluatedAt,
	}
}

type timerSessionModel struct {
	ID          uuid.UUID  `gorm:"type:uuid;default:gen_random_uuid();primaryKey"`
	UserID      uuid.UUID  `gorm:"column:user_id"`
	ProblemID   *uuid.UUID `gorm:"column:problem_id"`
	StartedAt   time.Time  `gorm:"column:started_at"`
	EndedAt     *time.Time `gorm:"column:ended_at"`
	ElapsedSecs int        `gorm:"column:elapsed_secs"`
}

func (timerSessionModel) TableName() string { return "timer_sessions" }

func (m timerSessionModel) toDomain() domain.TimerSession {
	return domain.TimerSession{
		ID:          m.ID,
		UserID:      m.UserID,
		ProblemID:   m.ProblemID,
		StartedAt:   m.StartedAt,
		EndedAt:     m.EndedAt,
		ElapsedSecs: m.ElapsedSecs,
	}
}

type dailySessionModel struct {
	ID             uuid.UUID `gorm:"type:uuid;default:gen_random_uuid();primaryKey"`
	UserID         uuid.UUID `gorm:"column:user_id"`
	Date           time.Time `gorm:"column:date"`
	SolvedCount    int       `gorm:"column:solved_count"`
	AttemptedCount int       `gorm:"column:attempted_count"`
	TimeSpentSecs  int       `gorm:"column:time_spent_secs"`
	CreatedAt      time.Time `gorm:"column:created_at"`
	UpdatedAt      time.Time `gorm:"column:updated_at"`
}

func (dailySessionModel) TableName() string { return "daily_sessions" }

func (m dailySessionModel) toDomain() domain.DailySession {
	return domain.DailySession{
		ID:             m.ID,
		UserID:         m.UserID,
		Date:           m.Date,
		SolvedCount:    m.SolvedCount,
		AttemptedCount: m.AttemptedCount,
		TimeSpentSecs:  m.TimeSpentSecs,
		CreatedAt:      m.CreatedAt,
		UpdatedAt:      m.UpdatedAt,
	}
}

type testCaseModel struct {
	ID        uuid.UUID `gorm:"type:uuid;default:gen_random_uuid();primaryKey"`
	ProblemID uuid.UUID `gorm:"column:problem_id"`
	Input     string    `gorm:"column:input"`
	Expected  string    `gorm:"column:expected"`
	IsHidden  bool      `gorm:"column:is_hidden"`
	OrderIdx  int       `gorm:"column:order_idx"`
}

func (testCaseModel) TableName() string { return "test_cases" }

func (m testCaseModel) toDomain() domain.TestCase {
	return domain.TestCase{
		ID:        m.ID,
		ProblemID: m.ProblemID,
		Input:     m.Input,
		Expected:  m.Expected,
		IsHidden:  m.IsHidden,
		OrderIdx:  m.OrderIdx,
	}
}

type registryVersionModel struct {
	ID        int64     `gorm:"column:id;primaryKey"`
	Version   string    `gorm:"column:version"`
	UpdatedAt time.Time `gorm:"column:updated_at"`
	Manifests []byte    `gorm:"column:manifests"`
	SyncedAt  time.Time `gorm:"column:synced_at"`
}

func (registryVersionModel) TableName() string { return "registry_versions" }

type reviewScheduleModel struct {
	ID           uuid.UUID `gorm:"type:uuid;default:gen_random_uuid();primaryKey"`
	UserID       uuid.UUID `gorm:"column:user_id"`
	ProblemID    uuid.UUID `gorm:"column:problem_id"`
	NextReviewAt time.Time `gorm:"column:next_review_at"`
	IntervalDays int       `gorm:"column:interval_days"`
	ReviewCount  int       `gorm:"column:review_count"`
	CreatedAt    time.Time `gorm:"column:created_at"`
	UpdatedAt    time.Time `gorm:"column:updated_at"`
}

func (reviewScheduleModel) TableName() string { return "review_schedules" }

type dailyMissionModel struct {
	ID            uuid.UUID `gorm:"type:uuid;default:gen_random_uuid();primaryKey"`
	UserID        uuid.UUID `gorm:"column:user_id"`
	Date          time.Time `gorm:"column:date"`
	RequiredTasks []byte    `gorm:"column:required_tasks"`
	OptionalTasks []byte    `gorm:"column:optional_tasks"`
	ReviewTasks   []byte    `gorm:"column:review_tasks"`
	GeneratedAt   time.Time `gorm:"column:generated_at"`
}

func (dailyMissionModel) TableName() string { return "daily_missions" }

func missionModelFromDomain(m domain.DailyMission) (dailyMissionModel, error) {
	req, err := json.Marshal(m.RequiredTasks)
	if err != nil {
		return dailyMissionModel{}, err
	}
	opt, err := json.Marshal(m.OptionalTasks)
	if err != nil {
		return dailyMissionModel{}, err
	}
	rev, err := json.Marshal(m.ReviewTasks)
	if err != nil {
		return dailyMissionModel{}, err
	}

	return dailyMissionModel{
		ID:            m.ID,
		UserID:        m.UserID,
		Date:          m.Date,
		RequiredTasks: req,
		OptionalTasks: opt,
		ReviewTasks:   rev,
		GeneratedAt:   m.GeneratedAt,
	}, nil
}
