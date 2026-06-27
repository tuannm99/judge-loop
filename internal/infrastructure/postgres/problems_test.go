package postgres

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
	"github.com/tuannm99/judge-loop/internal/domain"
)

func TestProblemRepositoryListAppliesFiltersAndPagination(t *testing.T) {
	db, mock := newMockDB(t)
	repository := NewProblemRepositoryImpl(db)
	difficulty := domain.DifficultyEasy
	provider := domain.ProviderLeetCode
	problemID := uuid.New()
	now := time.Now().UTC()

	mock.ExpectQuery(
		`SELECT COUNT\(DISTINCT\("problems"\."id"\)\) FROM "problems" `+
			`WHERE LOWER\(title\) LIKE \$1 AND difficulty = \$2 AND provider = \$3`,
	).
		WithArgs("%two sum%", string(difficulty), string(provider)).
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(1))
	mock.ExpectQuery(
		`SELECT DISTINCT problems\.\* FROM "problems" `+
			`WHERE LOWER\(title\) LIKE \$1 AND difficulty = \$2 AND provider = \$3 `+
			`ORDER BY created_at DESC LIMIT \$4 OFFSET \$5`,
	).
		WithArgs("%two sum%", string(difficulty), string(provider), 5, 10).
		WillReturnRows(sqlmock.NewRows([]string{
			"id",
			"slug",
			"title",
			"difficulty",
			"provider",
			"external_id",
			"source_url",
			"estimated_time",
			"description_markdown",
			"starter_code",
			"execution_spec",
			"judge_ready",
			"created_at",
			"updated_at",
		}).AddRow(
			problemID,
			"two-sum",
			"Two Sum",
			string(difficulty),
			string(provider),
			"1",
			"https://leetcode.com/problems/two-sum",
			15,
			"Find two numbers.",
			[]byte(`{"go":"package main"}`),
			[]byte(`{}`),
			true,
			now,
			now,
		))
	mock.ExpectQuery(
		`SELECT pll\.problem_id, pl\.slug FROM problem_label_links pll ` +
			`JOIN problem_labels pl ON pl\.id = pll\.problem_label_id ` +
			`WHERE pll\.problem_id IN \(\$1\) ORDER BY pl\.slug ASC`,
	).
		WithArgs(problemID).
		WillReturnRows(sqlmock.NewRows([]string{"problem_id", "slug"}).
			AddRow(problemID, "array"))

	problems, total, err := repository.List(context.Background(), ProblemFilter{
		Title:      "  TWO SUM ",
		Difficulty: &difficulty,
		Provider:   &provider,
		Limit:      5,
		Offset:     10,
	})

	require.NoError(t, err)
	require.Equal(t, 1, total)
	require.Len(t, problems, 1)
	require.Equal(t, "two-sum", problems[0].Slug)
	require.Equal(t, []string{"array"}, problems[0].Tags)
	require.Equal(t, "package main", problems[0].StarterCode["go"])
}

func TestProblemRepositoryListWrapsCountError(t *testing.T) {
	db, mock := newMockDB(t)
	repository := NewProblemRepositoryImpl(db)

	mock.ExpectQuery(`SELECT COUNT\(DISTINCT\("problems"\."id"\)\) FROM "problems"`).
		WillReturnError(errors.New("database unavailable"))

	problems, total, err := repository.List(context.Background(), ProblemFilter{})

	require.Nil(t, problems)
	require.Zero(t, total)
	require.ErrorContains(t, err, "count problems: database unavailable")
}

func TestProblemRepositoryLabelCRUD(t *testing.T) {
	db, mock := newMockDB(t)
	repository := NewProblemRepositoryImpl(db)
	labelID := uuid.New()
	now := time.Now().UTC()

	mock.ExpectQuery(`SELECT "slug" FROM "problem_labels" ORDER BY slug ASC`).
		WillReturnRows(sqlmock.NewRows([]string{"slug"}).AddRow("array").AddRow("hash-map"))
	labels, err := repository.ListLabels(context.Background(), "tag")
	require.NoError(t, err)
	require.Equal(t, []string{"array", "hash-map"}, labels)

	mock.ExpectQuery(`SELECT \* FROM "problem_labels" ORDER BY name ASC, slug ASC`).
		WillReturnRows(sqlmock.NewRows([]string{
			"id", "slug", "name", "created_at", "updated_at",
		}).AddRow(labelID, "array", "Array", now, now))
	records, err := repository.ListLabelRecords(context.Background(), "tag")
	require.NoError(t, err)
	require.Len(t, records, 1)
	require.Equal(t, "array", records[0].Slug)

	mock.ExpectQuery(`INSERT INTO "problem_labels".*RETURNING "id"`).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(labelID))
	created, err := repository.CreateLabel(context.Background(), domain.ProblemLabel{
		Slug: "array",
		Name: "Array",
	})
	require.NoError(t, err)
	require.Equal(t, labelID, created.ID)

	mock.ExpectExec(`UPDATE "problem_labels" SET .* WHERE id = \$[0-9]+`).
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectQuery(`SELECT \* FROM "problem_labels" WHERE id = \$1 ORDER BY "problem_labels"\."id" LIMIT \$2`).
		WithArgs(labelID, 1).
		WillReturnRows(sqlmock.NewRows([]string{"id", "slug", "name"}).
			AddRow(labelID, "arrays", "Arrays"))
	updated, err := repository.UpdateLabel(context.Background(), domain.ProblemLabel{
		ID:   labelID,
		Slug: "arrays",
		Name: "Arrays",
	})
	require.NoError(t, err)
	require.Equal(t, "arrays", updated.Slug)

	mock.ExpectExec(`DELETE FROM "problem_labels" WHERE id = \$1`).
		WithArgs(labelID).
		WillReturnResult(sqlmock.NewResult(0, 1))
	require.NoError(t, repository.DeleteLabel(context.Background(), labelID))
}

func TestProblemRepositoryLookupAndSuggestionHandleNotFound(t *testing.T) {
	db, mock := newMockDB(t)
	repository := NewProblemRepositoryImpl(db)
	problemID := uuid.New()
	userID := uuid.New()

	mock.ExpectQuery(`SELECT \* FROM "problems" WHERE id = \$1 ORDER BY "problems"\."id" LIMIT \$2`).
		WithArgs(problemID, 1).
		WillReturnRows(sqlmock.NewRows([]string{"id"}))
	byID, err := repository.GetByID(context.Background(), problemID)
	require.NoError(t, err)
	require.Nil(t, byID)

	mock.ExpectQuery(`SELECT \* FROM "problems" WHERE slug = \$1 ORDER BY "problems"\."id" LIMIT \$2`).
		WithArgs("missing", 1).
		WillReturnRows(sqlmock.NewRows([]string{"id"}))
	bySlug, err := repository.GetBySlug(context.Background(), "missing")
	require.NoError(t, err)
	require.Nil(t, bySlug)

	mock.ExpectQuery(`(?s)SELECT \* FROM "problems".*ORDER BY RANDOM\(\) LIMIT \$[0-9]+`).
		WillReturnRows(sqlmock.NewRows([]string{"id"}))
	suggested, err := repository.Suggest(context.Background(), userID, nil)
	require.NoError(t, err)
	require.Nil(t, suggested)
}

func TestProblemRepositoryMutationAndUnsolvedPaths(t *testing.T) {
	t.Run("upsert wraps insert error", func(t *testing.T) {
		db, mock := newMockDB(t)
		repository := NewProblemRepositoryImpl(db)
		mock.ExpectQuery(`INSERT INTO "problems".*ON CONFLICT.*RETURNING "id"`).
			WillReturnError(errors.New("database unavailable"))

		err := repository.UpsertFromManifest(context.Background(), domain.ProblemManifest{
			Provider:   domain.ProviderLeetCode,
			ExternalID: "1",
			Slug:       "two-sum",
			Title:      "Two Sum",
		})
		require.ErrorContains(t, err, "upsert problem leetcode/two-sum")
	})

	t.Run("update rolls back on update error", func(t *testing.T) {
		db, mock := newMockDB(t)
		repository := NewProblemRepositoryImpl(db)
		problemID := uuid.New()
		mock.ExpectBegin()
		mock.ExpectExec(`UPDATE "problems" SET .* WHERE id = \$[0-9]+`).
			WillReturnError(errors.New("database unavailable"))
		mock.ExpectRollback()

		_, err := repository.Update(context.Background(), problemID, domain.ProblemManifest{
			Slug:  "two-sum",
			Title: "Two Sum",
		})
		require.ErrorContains(t, err, "update problem")
	})

	t.Run("unsolved returns empty result", func(t *testing.T) {
		db, mock := newMockDB(t)
		repository := NewProblemRepositoryImpl(db)
		mock.ExpectQuery(`(?s)SELECT \* FROM "problems".*ORDER BY RANDOM\(\) LIMIT \$[0-9]+`).
			WillReturnRows(sqlmock.NewRows([]string{"id"}))

		problems, err := repository.GetUnsolved(context.Background(), uuid.New(), 0)
		require.NoError(t, err)
		require.Empty(t, problems)
	})
}

func TestProblemRepositoryTagFilterAndLabelReplacement(t *testing.T) {
	t.Run("tag filter is applied to count", func(t *testing.T) {
		db, mock := newMockDB(t)
		repository := NewProblemRepositoryImpl(db)
		mock.ExpectQuery(`(?s)SELECT COUNT.*problem_label_links.*HAVING COUNT\(DISTINCT pl\.slug\) = \$[0-9]+`).
			WillReturnError(errors.New("stop after query assertion"))

		_, _, err := repository.List(context.Background(), ProblemFilter{
			Tags: []string{"array", "hash-map"},
		})
		require.ErrorContains(t, err, "stop after query assertion")
	})

	t.Run("empty labels replace existing links transactionally", func(t *testing.T) {
		db, mock := newMockDB(t)
		repository := NewProblemRepositoryImpl(db)
		problemID := uuid.New()
		mock.ExpectBegin()
		mock.ExpectExec(`SAVEPOINT sp[0-9]+`).
			WillReturnResult(sqlmock.NewResult(0, 0))
		mock.ExpectExec(`DELETE FROM "problem_label_links" WHERE problem_id = \$1`).
			WithArgs(problemID).
			WillReturnResult(sqlmock.NewResult(0, 2))
		mock.ExpectCommit()

		require.NoError(t, repository.replaceProblemLabels(context.Background(), problemID, nil))
	})
}
