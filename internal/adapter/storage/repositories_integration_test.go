package storageadapter

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/lib/pq"
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
	"github.com/tuannm99/judge-loop/internal/domain"
	postgres "github.com/tuannm99/judge-loop/internal/infrastructure/postgres"
	outport "github.com/tuannm99/judge-loop/internal/port/out"
)

func newStorageTestDB(t *testing.T) *postgres.DB {
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

	var db *postgres.DB
	require.Eventually(t, func() bool {
		db, err = postgres.Connect(context.Background(), dsn)
		return err == nil
	}, 30*time.Second, 500*time.Millisecond)
	t.Cleanup(db.Close)
	return db
}

func seedStorageUser(t *testing.T, db *postgres.DB) uuid.UUID {
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

func seedStorageProblem(t *testing.T, db *postgres.DB, slug, extID string, tags, patterns []string) uuid.UUID {
	t.Helper()
	id := uuid.New()
	require.NoError(t, db.Gorm.Exec(
		`INSERT INTO problems (id, slug, title, difficulty, tags, pattern_tags, provider, external_id, source_url, estimated_time)
		 VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		id, slug, "Title "+slug, "easy", pq.StringArray(tags), pq.StringArray(patterns), "leetcode", extID, "https://example.com/"+slug, 15,
	).Error)
	return id
}

func TestStorageRepositoriesIntegration(t *testing.T) {
	db := newStorageTestDB(t)
	ctx := context.Background()

	problems := NewProblemRepository(db)
	submissions := NewSubmissionRepository(db)
	sessions := NewSessionRepository(db)
	reviews := NewReviewRepository(db)
	registry := NewRegistryRepository(db)
	testCases := NewTestCaseRepository(db)

	userID := seedStorageUser(t, db)
	problemID := seedStorageProblem(t, db, "two-sum", "ext-1", []string{"array"}, []string{"hash-map"})
	otherProblemID := seedStorageProblem(t, db, "best-time", "ext-2", []string{"array"}, []string{"sliding-window"})

	problemRepo := NewProblemRepository(db)
	require.NotNil(t, problemRepo)
	diff := domain.DifficultyEasy
	prov := domain.ProviderLeetCode
	listed, total, err := problems.List(ctx, outport.ProblemFilter{
		Difficulty: &diff,
		Provider:   &prov,
		Tag:        "array",
		Pattern:    "hash-map",
		Limit:      10,
	})
	require.NoError(t, err)
	require.Equal(t, 1, total)
	require.Len(t, listed, 1)

	gotProblem, err := problems.GetByID(ctx, problemID)
	require.NoError(t, err)
	require.NotNil(t, gotProblem)

	gotProblem, err = problems.GetBySlug(ctx, "two-sum")
	require.NoError(t, err)
	require.NotNil(t, gotProblem)

	suggested, err := problems.Suggest(ctx, userID, []string{"hash-map"})
	require.NoError(t, err)
	require.NotNil(t, suggested)

	require.NoError(t, problems.UpsertFromManifest(ctx, domain.ProblemManifest{
		Slug:          "dynamic-programming",
		Title:         "DP",
		Difficulty:    domain.DifficultyMedium,
		Tags:          []string{"dp"},
		PatternTags:   []string{"dp"},
		Provider:      domain.ProviderLeetCode,
		ExternalID:    "ext-3",
		SourceURL:     "https://example.com/dp",
		EstimatedTime: 30,
	}))

	sessionRepo := NewSessionRepository(db)
	require.NotNil(t, sessionRepo)
	daily, err := sessions.GetOrCreateToday(ctx, userID)
	require.NoError(t, err)
	require.NotNil(t, daily)
	require.NoError(t, sessions.RecordSubmission(ctx, userID, false))

	timer, err := sessions.StartTimer(ctx, userID, &problemID)
	require.NoError(t, err)
	require.NotNil(t, timer)
	active, err := sessions.ActiveTimer(ctx, userID)
	require.NoError(t, err)
	require.NotNil(t, active)
	require.GreaterOrEqual(t, active.StartedAt.Unix(), timer.StartedAt.Unix())
	time.Sleep(1100 * time.Millisecond)
	stopped, err := sessions.StopTimer(ctx, userID)
	require.NoError(t, err)
	require.NotNil(t, stopped)
	streak, err := sessions.GetStreak(ctx, userID)
	require.NoError(t, err)
	require.Equal(t, 0, streak.Current)

	submissionRepo := NewSubmissionRepository(db)
	require.NotNil(t, submissionRepo)
	sub := &domain.Submission{
		UserID:    userID,
		ProblemID: problemID,
		SessionID: &timer.ID,
		Language:  domain.LanguageGo,
		Code:      "package main\nfunc main(){}",
	}
	require.NoError(t, submissions.Create(ctx, sub))
	now := time.Now().UTC()
	require.NoError(t, submissions.UpdateVerdict(ctx, sub.ID, string(domain.StatusAccepted), string(domain.VerdictAccepted), 1, 1, 5, "", &now))
	gotSub, err := submissions.GetByID(ctx, sub.ID)
	require.NoError(t, err)
	require.NotNil(t, gotSub)
	listSubs, err := submissions.ListByUser(ctx, userID, &problemID, 20, 0)
	require.NoError(t, err)
	require.Len(t, listSubs, 1)

	require.NoError(t, sessions.RecordSubmission(ctx, userID, true))
	streak, err = sessions.GetStreak(ctx, userID)
	require.NoError(t, err)
	require.Equal(t, 1, streak.Current)

	require.NoError(t, db.Gorm.Exec(
		`INSERT INTO test_cases (id, problem_id, input, expected, is_hidden, order_idx) VALUES (?, ?, ?, ?, ?, ?)`,
		uuid.New(), problemID, "1 2", "3", false, 1,
	).Error)
	require.NoError(t, db.Gorm.Exec(
		`INSERT INTO test_cases (id, problem_id, input, expected, is_hidden, order_idx) VALUES (?, ?, ?, ?, ?, ?)`,
		uuid.New(), problemID, "2 3", "5", true, 2,
	).Error)
	tcRepo := NewTestCaseRepository(db)
	require.NotNil(t, tcRepo)
	cases, err := testCases.GetByProblem(ctx, problemID)
	require.NoError(t, err)
	require.Len(t, cases, 1)

	reviewRepo := NewReviewRepository(db)
	require.NotNil(t, reviewRepo)
	require.NoError(t, reviews.Upsert(ctx, userID, otherProblemID))
	var schedule struct {
		ID           uuid.UUID
		NextReviewAt time.Time
	}
	require.NoError(t, db.Gorm.Table("review_schedules").Select("id, next_review_at").Where("user_id = ? AND problem_id = ?", userID, otherProblemID).Take(&schedule).Error)
	require.NoError(t, db.Gorm.Exec(`UPDATE review_schedules SET next_review_at = ? WHERE id = ?`, time.Now().Add(-24*time.Hour), schedule.ID).Error)
	due, err := reviews.GetDue(ctx, userID)
	require.NoError(t, err)
	require.Len(t, due, 1)

	regRepo := NewRegistryRepository(db)
	require.NotNil(t, regRepo)
	row, err := registry.GetLatest(ctx)
	require.NoError(t, err)
	require.Nil(t, row)
	refs := []domain.ManifestRef{{Name: "main", Path: "providers/main.json"}}
	require.NoError(t, registry.Save(ctx, "v1", now, refs))
	row, err = registry.GetLatest(ctx)
	require.NoError(t, err)
	require.NotNil(t, row)
	require.Equal(t, "v1", row.Version)
}
