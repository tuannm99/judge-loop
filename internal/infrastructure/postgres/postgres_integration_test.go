package postgres

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
	"github.com/tuannm99/judge-loop/internal/domain"
)

func newIntegrationDB(t *testing.T) *DB {
	t.Helper()

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()

	container, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: testcontainers.ContainerRequest{
			Image:        "postgres:16-alpine",
			ExposedPorts: []string{"5432/tcp"},
			Env: map[string]string{
				"POSTGRES_USER":     "judge",
				"POSTGRES_PASSWORD": "judge",
				"POSTGRES_DB":       "judge",
			},
			WaitingFor: wait.ForListeningPort("5432/tcp"),
		},
		Started: true,
	})
	if err != nil {
		t.Skipf("docker/testcontainers unavailable: %v", err)
	}
	t.Cleanup(func() {
		_ = container.Terminate(context.Background())
	})

	host, err := container.Host(ctx)
	require.NoError(t, err)
	port, err := container.MappedPort(ctx, "5432/tcp")
	require.NoError(t, err)

	dsn := fmt.Sprintf(
		"host=%s port=%s user=judge password=judge dbname=judge sslmode=disable",
		host,
		port.Port(),
	)

	var db *DB
	require.Eventually(t, func() bool {
		db, err = Connect(context.Background(), dsn)
		return err == nil
	}, 30*time.Second, 500*time.Millisecond, "connect postgres")

	require.NoError(t, RunMigrations(context.Background(), db.SQL), "migrate postgres")

	t.Cleanup(func() {
		db.Close()
	})

	return db
}

func seedUser(t *testing.T, db *DB) uuid.UUID {
	t.Helper()

	id := uuid.New()
	require.NoError(t, db.Gorm.Exec(
		`INSERT INTO users (id, username, email) VALUES (?, ?, ?)`,
		id,
		"user-"+id.String()[:8],
		id.String()+"@example.com",
	).Error)
	return id
}

func seedProblem(t *testing.T, db *DB, slug string, tags []string) uuid.UUID {
	t.Helper()

	store := NewProblemRepositoryImpl(db)
	manifest := domain.ProblemManifest{
		Provider:      domain.ProviderLeetCode,
		ExternalID:    slug,
		Slug:          slug,
		Title:         "Title " + slug,
		Difficulty:    domain.DifficultyEasy,
		Tags:          tags,
		SourceURL:     "https://example.com/" + slug,
		EstimatedTime: 15,
		StarterCode:   map[string]string{},
	}
	require.NoError(t, store.UpsertFromManifest(context.Background(), manifest))

	problem, err := store.GetBySlug(context.Background(), slug)
	require.NoError(t, err)
	require.NotNil(t, problem)
	return problem.ID
}

func TestConnectRegistryAndMissionIntegration(t *testing.T) {
	db := newIntegrationDB(t)
	ctx := context.Background()

	registryStore := NewRegistryRepositoryImpl(db)
	latest, err := registryStore.GetLatest(ctx)
	require.NoError(t, err)
	require.Nil(t, latest)
	require.True(t, registryStore.LastSyncedAt(ctx).IsZero())

	updatedAt := time.Date(2026, 3, 31, 9, 0, 0, 0, time.UTC)
	refs := []domain.ManifestRef{{Name: "main", Path: "providers/main.json"}}
	require.NoError(t, registryStore.Save(ctx, "v1", updatedAt, refs))

	latest, err = registryStore.GetLatest(ctx)
	require.NoError(t, err)
	require.NotNil(t, latest)
	require.Equal(t, "v1", latest.Version)
	require.Equal(t, updatedAt, latest.UpdatedAt.UTC())
	require.False(t, registryStore.LastSyncedAt(ctx).IsZero())

	userID := seedUser(t, db)
	missionStore := NewMissionRepositoryImpl(db)
	current, err := missionStore.GetToday(ctx, userID)
	require.NoError(t, err)
	require.Nil(t, current)

	mission := domain.DailyMission{
		ID:     uuid.New(),
		UserID: userID,
		Date:   time.Now().UTC().Truncate(24 * time.Hour),
		RequiredTasks: []domain.MissionTask{
			{ProblemID: uuid.New(), Slug: "two-sum", Title: "Two Sum", Reason: "unsolved", Priority: 3},
		},
		OptionalTasks: []domain.MissionTask{
			{ProblemID: uuid.New(), Slug: "reverse", Title: "Reverse", Reason: "optional challenge", Priority: 1},
		},
		ReviewTasks: []domain.MissionTask{
			{ProblemID: uuid.New(), Slug: "dp", Title: "DP", Reason: "due for review", Priority: 10},
		},
		GeneratedAt: time.Now().UTC(),
	}
	require.NoError(t, missionStore.Save(ctx, mission))

	current, err = missionStore.GetToday(ctx, userID)
	require.NoError(t, err)
	require.NotNil(t, current)
	require.Len(t, current.RequiredTasks, 1)
	require.Len(t, current.OptionalTasks, 1)
	require.Len(t, current.ReviewTasks, 1)
}

func TestProblemSubmissionSessionAndPerformanceIntegration(t *testing.T) {
	db := newIntegrationDB(t)
	ctx := context.Background()
	userID := seedUser(t, db)

	problemStore := NewProblemRepositoryImpl(db)
	submissionStore := NewSubmissionRepositoryImpl(db)
	sessionStore := NewSessionRepositoryImpl(db)
	testCaseStore := NewTestCaseRepositoryImpl(db)
	performanceStore := NewPerformanceRepositoryImpl(db)

	problemA := domain.ProblemManifest{
		Slug:                "two-sum",
		Title:               "Two Sum",
		Difficulty:          domain.DifficultyEasy,
		Tags:                []string{"array", "hash-map"},
		Provider:            domain.ProviderLeetCode,
		ExternalID:          "1",
		SourceURL:           "https://example.com/two-sum",
		EstimatedTime:       15,
		DescriptionMarkdown: "## Two Sum\n\nReturn indices for the matching pair.",
		StarterCode: map[string]string{
			"python": "class Solution:\n    pass\n",
			"go":     "package main\n\nfunc main() {}\n",
		},
	}
	problemB := domain.ProblemManifest{
		Slug:          "best-time",
		Title:         "Best Time",
		Difficulty:    domain.DifficultyEasy,
		Tags:          []string{"array", "sliding-window"},
		Provider:      domain.ProviderLeetCode,
		ExternalID:    "2",
		SourceURL:     "https://example.com/best-time",
		EstimatedTime: 20,
	}
	require.NoError(t, problemStore.UpsertFromManifest(ctx, problemA))
	require.NoError(t, problemStore.UpsertFromManifest(ctx, problemB))

	problems, total, err := problemStore.List(ctx, ProblemFilter{Tags: []string{"array", "hash-map"}, Limit: 10})
	require.NoError(t, err)
	require.Equal(t, 1, total)
	require.Len(t, problems, 1)

	first := problems[0]
	byID, err := problemStore.GetByID(ctx, first.ID)
	require.NoError(t, err)
	require.NotNil(t, byID)
	require.Equal(t, first.Slug, byID.Slug)

	bySlug, err := problemStore.GetBySlug(ctx, "two-sum")
	require.NoError(t, err)
	require.NotNil(t, bySlug)
	require.Equal(t, "two-sum", bySlug.Slug)
	require.Equal(t, "## Two Sum\n\nReturn indices for the matching pair.", bySlug.DescriptionMarkdown)
	require.Equal(t, "class Solution:\n    pass\n", bySlug.StarterCode["python"])
	require.Equal(t, "package main\n\nfunc main() {}\n", bySlug.StarterCode["go"])

	suggested, err := problemStore.Suggest(ctx, userID, []string{"hash-map"})
	require.NoError(t, err)
	require.NotNil(t, suggested)

	require.NoError(t, db.Gorm.Create(&testCaseModel{
		ID:        uuid.New(),
		ProblemID: first.ID,
		Input:     "1 2",
		Expected:  "3",
		IsHidden:  false,
		OrderIdx:  1,
	}).Error)
	require.NoError(t, db.Gorm.Create(&testCaseModel{
		ID:        uuid.New(),
		ProblemID: first.ID,
		Input:     "2 3",
		Expected:  "5",
		IsHidden:  true,
		OrderIdx:  2,
	}).Error)
	testCases, err := testCaseStore.GetByProblem(ctx, first.ID)
	require.NoError(t, err)
	require.Len(t, testCases, 1)

	daily, err := sessionStore.GetOrCreateToday(ctx, userID)
	require.NoError(t, err)
	require.NotNil(t, daily)

	timer, err := sessionStore.StartTimer(ctx, userID, &first.ID)
	require.NoError(t, err)
	require.NotNil(t, timer)

	active, err := sessionStore.ActiveTimer(ctx, userID)
	require.NoError(t, err)
	require.NotNil(t, active)
	require.Equal(t, timer.ID, active.ID)

	time.Sleep(1100 * time.Millisecond)
	stopped, err := sessionStore.StopTimer(ctx, userID)
	require.NoError(t, err)
	require.NotNil(t, stopped)
	require.GreaterOrEqual(t, stopped.ElapsedSecs, 1)

	sub := &domain.Submission{
		UserID:    userID,
		ProblemID: first.ID,
		SessionID: &timer.ID,
		Language:  domain.LanguageGo,
		Code:      "package main\nfunc main(){}",
	}
	require.NoError(t, submissionStore.Create(ctx, sub))
	require.NotEqual(t, uuid.Nil, sub.ID)
	require.Equal(t, domain.StatusPending, sub.Status)

	now := time.Now().UTC()
	require.NoError(t, submissionStore.UpdateVerdict(
		ctx,
		sub.ID,
		string(domain.StatusAccepted),
		string(domain.VerdictAccepted),
		1,
		1,
		5,
		"",
		&now,
	))
	require.NoError(t, sessionStore.RecordSubmission(ctx, userID, true))

	gotSub, err := submissionStore.GetByID(ctx, sub.ID)
	require.NoError(t, err)
	require.NotNil(t, gotSub)
	require.Equal(t, domain.StatusAccepted, gotSub.Status)

	listed, err := submissionStore.ListByUser(ctx, userID, &first.ID, 20, 0)
	require.NoError(t, err)
	require.Len(t, listed, 1)

	unsolved, err := problemStore.GetUnsolved(ctx, userID, 10)
	require.NoError(t, err)
	require.Len(t, unsolved, 1)
	require.Equal(t, "best-time", unsolved[0].Slug)

	suggested, err = problemStore.Suggest(ctx, userID, []string{"hash-map"})
	require.NoError(t, err)
	require.NotNil(t, suggested)
	require.NotEqual(t, "two-sum", suggested.Slug)

	patternScores, err := performanceStore.GetPatternScores(ctx, userID)
	require.NoError(t, err)
	require.Equal(t, 1.0, patternScores["hash-map"])

	stats, err := performanceStore.GetStats(ctx, userID)
	require.NoError(t, err)
	require.Equal(t, 1, stats.TotalAttempts)
	require.Equal(t, 1, stats.AcceptedCount)
	require.GreaterOrEqual(t, stats.AvgSolveTime, 0.0)

	streak, err := sessionStore.GetStreak(ctx, userID)
	require.NoError(t, err)
	require.Equal(t, 1, streak.Current)
	require.Equal(t, 1, streak.Longest)
	require.NotNil(t, streak.LastPracticed)

	require.Equal(t, stopped.ElapsedSecs, ElapsedNow(stopped))
	require.Nil(t, active.EndedAt)
	require.GreaterOrEqual(t, ElapsedNow(active), 0)
}

func TestReviewStoreIntegration(t *testing.T) {
	db := newIntegrationDB(t)
	ctx := context.Background()
	userID := seedUser(t, db)
	problemID := seedProblem(t, db, "review-me", []string{"array", "dp"})

	reviewStore := NewReviewRepositoryImpl(db)

	due, err := reviewStore.GetDue(ctx, userID)
	require.NoError(t, err)
	require.Empty(t, due)

	require.NoError(t, reviewStore.Upsert(ctx, userID, problemID))

	var schedule reviewScheduleModel
	require.NoError(t, db.Gorm.Where("user_id = ? AND problem_id = ?", userID, problemID).Take(&schedule).Error)
	require.Equal(t, 1, schedule.IntervalDays)
	require.Equal(t, 1, schedule.ReviewCount)

	schedule.NextReviewAt = time.Now().Add(-24 * time.Hour)
	require.NoError(t, db.Gorm.Save(&schedule).Error)
	due, err = reviewStore.GetDue(ctx, userID)
	require.NoError(t, err)
	require.Len(t, due, 1)
	require.Equal(t, problemID, due[0].ProblemID)

	require.NoError(t, reviewStore.Upsert(ctx, userID, problemID))
	require.NoError(t, db.Gorm.Where("user_id = ? AND problem_id = ?", userID, problemID).Take(&schedule).Error)
	require.Equal(t, 3, schedule.IntervalDays)
	require.Equal(t, 2, schedule.ReviewCount)
}
