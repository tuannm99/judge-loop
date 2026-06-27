package postgres

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
	"github.com/tuannm99/judge-loop/internal/domain"
)

func TestTestCaseRepositoryGetByProblemMapsMetadata(t *testing.T) {
	db, mock := newMockDB(t)
	repository := NewTestCaseRepositoryImpl(db)
	problemID := uuid.New()
	testCaseID := uuid.New()

	mock.ExpectQuery(
		`SELECT \* FROM "test_cases" WHERE problem_id = \$1 AND is_hidden = false ORDER BY order_idx`,
	).
		WithArgs(problemID).
		WillReturnRows(sqlmock.NewRows([]string{
			"id", "problem_id", "name", "input", "expected", "input_json",
			"expected_json", "metadata", "is_hidden", "order_idx",
		}).AddRow(
			testCaseID,
			problemID,
			"example",
			"2 7 9",
			"0 1",
			[]byte(`{"args":[[2,7],9]}`),
			[]byte(`[0,1]`),
			[]byte(`{"source":"example"}`),
			false,
			0,
		))

	testCases, err := repository.GetByProblem(context.Background(), problemID)
	require.NoError(t, err)
	require.Len(t, testCases, 1)
	require.Equal(t, "example", testCases[0].Name)
	require.JSONEq(t, `{"source":"example"}`, string(testCases[0].Metadata))
}

func TestTestCaseRepositoryReplaceEmptySetUsesTransaction(t *testing.T) {
	db, mock := newMockDB(t)
	repository := NewTestCaseRepositoryImpl(db)
	problemID := uuid.New()

	mock.ExpectBegin()
	mock.ExpectExec(`DELETE FROM "test_cases" WHERE problem_id = \$1`).
		WithArgs(problemID).
		WillReturnResult(sqlmock.NewResult(0, 2))
	mock.ExpectCommit()

	require.NoError(t, repository.ReplaceForProblem(context.Background(), problemID, nil))
}

func TestTestCaseRepositoryGetAllAndReplace(t *testing.T) {
	db, mock := newMockDB(t)
	repository := NewTestCaseRepositoryImpl(db)
	problemID := uuid.New()

	mock.ExpectQuery(`SELECT \* FROM "test_cases" WHERE problem_id = \$1 ORDER BY order_idx`).
		WithArgs(problemID).
		WillReturnRows(sqlmock.NewRows([]string{
			"id", "problem_id", "input", "expected", "is_hidden", "order_idx",
		}).AddRow(uuid.New(), problemID, "input", "output", true, 0))
	all, err := repository.GetAllByProblem(context.Background(), problemID)
	require.NoError(t, err)
	require.Len(t, all, 1)
	require.True(t, all[0].IsHidden)

	mock.ExpectBegin()
	mock.ExpectExec(`DELETE FROM "test_cases" WHERE problem_id = \$1`).
		WithArgs(problemID).
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectQuery(`INSERT INTO "test_cases".*RETURNING "id"`).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(uuid.New()))
	mock.ExpectCommit()

	err = repository.ReplaceForProblem(context.Background(), problemID, []domain.TestCase{{
		Name:         "structured",
		InputJSON:    json.RawMessage(`{"args":[1]}`),
		ExpectedJSON: json.RawMessage(`2`),
	}})
	require.NoError(t, err)
}

func TestTestCaseJSONDefaults(t *testing.T) {
	require.Nil(t, nullableRawJSON(nil))
	require.JSONEq(t, `{"args":[1]}`, string(nullableRawJSON(json.RawMessage(`{"args":[1]}`))))
	require.JSONEq(t, `{}`, string(defaultRawJSON(nil)))
}
